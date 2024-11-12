// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package policy

import (
	"iter"
	"sync/atomic"

	"github.com/cilium/cilium/pkg/container/versioned"
	identityPkg "github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/identity/identitymanager"
	"github.com/cilium/cilium/pkg/lock"
)

// SelectorPolicy represents a cached selectorPolicy, previously resolved from
// the policy repository and ready to be distilled against a set of identities
// to compute datapath-level policy configuration.
type SelectorPolicy interface {
	// CreateRedirects is used to ensure the endpoint has created all the needed redirects
	// before a new EndpointPolicy is created.
	RedirectFilters() iter.Seq2[*L4Filter, *PerSelectorPolicy]

	// Consume returns the policy in terms of connectivity to peer
	// Identities.
	Consume(owner PolicyOwner, redirects map[string]uint16) *EndpointPolicy
}

// PolicyCache represents a cache of resolved policies for identities.
type PolicyCache struct {
	lock.Mutex

	// repo is a circular reference back to the Repository, but as
	// we create only one Repository and one PolicyCache for each
	// Cilium Agent process, these will never need to be garbage
	// collected.
	repo     *Repository
	policies map[identityPkg.NumericIdentity]*cachedSelectorPolicy
}

// NewPolicyCache creates a new cache of SelectorPolicy.
func NewPolicyCache(repo *Repository, idmgr identitymanager.IDManager) *PolicyCache {
	cache := &PolicyCache{
		repo:     repo,
		policies: make(map[identityPkg.NumericIdentity]*cachedSelectorPolicy),
	}
	if idmgr != nil {
		idmgr.Subscribe(cache)
	}
	return cache
}

// lookupOrCreate adds the specified Identity to the policy cache, with a reference
// from the specified Endpoint, then returns the threadsafe copy of the policy.
func (cache *PolicyCache) lookupOrCreate(identity *identityPkg.Identity) *cachedSelectorPolicy {
	cache.Lock()
	defer cache.Unlock()
	cip, ok := cache.policies[identity.ID]
	if !ok {
		cip = newCachedSelectorPolicy(identity)
		cache.policies[identity.ID] = cip
	}
	return cip
}

// delete forgets about any cached SelectorPolicy that this endpoint uses.
//
// Returns true if the SelectorPolicy was removed from the cache.
func (cache *PolicyCache) delete(identity *identityPkg.Identity) bool {
	cache.Lock()
	defer cache.Unlock()
	cip, ok := cache.policies[identity.ID]
	if ok {
		delete(cache.policies, identity.ID)
		cip.getPolicy().Detach()
	}
	return ok
}

// updateSelectorPolicy resolves the policy for the security identity of the
// specified endpoint and stores it internally. It will skip policy resolution
// if the cached policy is already at the revision specified in the repo.
//
// Returns whether the cache was updated, or an error.
//
// Must be called with repo.Mutex held for reading.
func (cache *PolicyCache) updateSelectorPolicy(identity *identityPkg.Identity) (*cachedSelectorPolicy, bool, error) {
	cip := cache.lookupOrCreate(identity)

	// As long as UpdatePolicy() is triggered from endpoint
	// regeneration, it's possible for two endpoints with the
	// *same* identity to race to update the policy here. Such
	// racing would lead to first of the endpoints using a
	// selectorPolicy that is already detached from the selector
	// cache, and thus not getting any incremental updates.
	//
	// Lock the 'cip' for the duration of the revision check and
	// the possible policy update.
	cip.Lock()
	defer cip.Unlock()

	// Don't resolve policy if it was already done for this or later revision.
	if policy := cip.getPolicy(); policy != nil && policy.Revision >= cache.repo.GetRevision() {
		return cip, false, nil
	}

	// Resolve the policies, which could fail
	selPolicy, err := cache.repo.resolvePolicyLocked(identity)
	if err != nil {
		return nil, false, err
	}

	cip.setPolicy(selPolicy)

	return cip, true, nil
}

// LocalEndpointIdentityAdded is not needed; we only care about local endpoint
// deletion
func (cache *PolicyCache) LocalEndpointIdentityAdded(identity *identityPkg.Identity) {
}

// LocalEndpointIdentityRemoved deletes the cached SelectorPolicy for the
// specified Identity.
func (cache *PolicyCache) LocalEndpointIdentityRemoved(identity *identityPkg.Identity) {
	cache.delete(identity)
}

// UpdatePolicy resolves the policy for the security identity of the specified
// endpoint and caches it for future use.
//
// The caller must provide threadsafety for iteration over the policy
// repository.
func (cache *PolicyCache) UpdatePolicy(identity *identityPkg.Identity) (SelectorPolicy, error) {
	cip, _, err := cache.updateSelectorPolicy(identity)
	return cip, err
}

// GetAuthTypes returns the AuthTypes required by the policy between the localID and remoteID, if
// any, otherwise returns nil.
func (cache *PolicyCache) GetAuthTypes(localID, remoteID identityPkg.NumericIdentity) AuthTypes {
	cache.Lock()
	cip, ok := cache.policies[localID]
	cache.Unlock()
	if !ok {
		return nil // No policy for localID (no endpoint with localID)
	}

	// SelectorPolicy is const after it has been created, so no locking needed to access it
	selPolicy := cip.getPolicy()

	var resTypes AuthTypes
	for cs, authTypes := range selPolicy.L4Policy.AuthMap {
		missing := false
		for authType := range authTypes {
			if _, exists := resTypes[authType]; !exists {
				missing = true
				break
			}
		}
		// Only check if 'cs' selects 'remoteID' if one of the authTypes is still missing
		// from the result
		if missing && cs.Selects(versioned.Latest(), remoteID) {
			if resTypes == nil {
				resTypes = make(AuthTypes, 1)
			}
			for authType := range authTypes {
				resTypes[authType] = struct{}{}
			}
		}
	}
	return resTypes
}

// cachedSelectorPolicy is a wrapper around a selectorPolicy (stored in the
// 'policy' field). It is always nested directly in the owning policyCache,
// and is protected against concurrent writes via the policyCache mutex.
type cachedSelectorPolicy struct {
	lock.Mutex // lock is needed to synchronize parallel policy updates

	identity *identityPkg.Identity
	policy   atomic.Pointer[selectorPolicy]
}

func newCachedSelectorPolicy(identity *identityPkg.Identity) *cachedSelectorPolicy {
	cip := &cachedSelectorPolicy{
		identity: identity,
	}
	return cip
}

// getPolicy returns a reference to the selectorPolicy that is cached.
//
// Users should treat the result as immutable state that MUST NOT be modified.
func (cip *cachedSelectorPolicy) getPolicy() *selectorPolicy {
	return cip.policy.Load()
}

// setPolicy updates the reference to the SelectorPolicy that is cached.
// Calls Detach() on the old policy, if any.
func (cip *cachedSelectorPolicy) setPolicy(policy *selectorPolicy) {
	oldPolicy := cip.policy.Swap(policy)
	if oldPolicy != nil {
		// Release the references the previous policy holds on the selector cache.
		oldPolicy.Detach()
	}
}

// Consume returns the EndpointPolicy that defines connectivity policy to
// Identities in the specified cache.
//
// This denotes that a particular endpoint is 'consuming' the policy from the
// selector policy cache.
func (cip *cachedSelectorPolicy) Consume(owner PolicyOwner, redirects map[string]uint16) *EndpointPolicy {
	// TODO: This currently computes the EndpointPolicy from SelectorPolicy
	// on-demand, however in future the cip is intended to cache the
	// EndpointPolicy for this Identity and emit datapath deltas instead.
	isHost := cip.identity.ID == identityPkg.ReservedIdentityHost
	return cip.getPolicy().DistillPolicy(owner, redirects, isHost)
}

func (cip *cachedSelectorPolicy) RedirectFilters() iter.Seq2[*L4Filter, *PerSelectorPolicy] {
	return cip.getPolicy().RedirectFilters()
}
