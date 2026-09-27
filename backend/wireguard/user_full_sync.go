package wireguard

import (
	"fmt"

	"github.com/pasarguard/node/common"
)

func (wg *WireGuard) syncUsersFull(users []*common.User) error {
	normalizedUsers := normalizeUsers(users)
	existingByKey := wg.buildExistingPeersByKeySnapshot()

	desiredPeers, err := wg.collectDesiredPeers(normalizedUsers)
	if err != nil {
		return err
	}

	diff, err := wg.buildSyncDiff(existingByKey, desiredPeers)
	if err != nil {
		return err
	}

	// A full sync is also the reconciliation boundary for interface-wide
	// peer state. In particular, the core PSK is not stored in PeerInfo, so a
	// PSK rotation can leave an otherwise identical peer looking unchanged.
	// AWG must still receive the authoritative peer snapshot in that case.
	psk, _ := wg.config.GetPreSharedKey()
	if !diff.Changed && psk == nil && !wg.config.AmneziaWG {
		return nil
	}
	peerConfigs, appliedKeys := wg.buildTargetPeerConfigs(diff.TargetPeers, psk)

	wg.mu.RLock()
	if err := wg.ensureRunningWithManagerLocked(); err != nil {
		wg.mu.RUnlock()
		return err
	}
	manager := wg.manager
	wg.mu.RUnlock()

	if err := manager.ApplyPeersReplaceAll(peerConfigs); err != nil {
		return fmt.Errorf("failed to apply full peer snapshot: %w", err)
	}

	var validTargetPeers []*PeerInfo
	for key, peer := range diff.TargetPeers {
		if _, ok := appliedKeys[key]; ok {
			validTargetPeers = append(validTargetPeers, peer)
		}
	}

	removedKeys := wg.peerStore.ReplaceAll(validTargetPeers)
	for _, key := range removedKeys {
		wg.statsTracker.RemoveStats(key)
	}

	return nil
}

func (wg *WireGuard) buildExistingPeersByKeySnapshot() map[string]*PeerInfo {
	existingPeers := wg.peerStore.GetAll()
	existingByKey := make(map[string]*PeerInfo, len(existingPeers))
	for _, peer := range existingPeers {
		existingByKey[peer.PublicKey.String()] = peer
	}

	return existingByKey
}
