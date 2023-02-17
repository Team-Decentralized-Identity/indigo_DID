package indexer

import (
	"context"
	"crypto"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/plc"
	did "github.com/whyrusleeping/go-did"
	"go.opentelemetry.io/otel"
)

type KeyManager struct {
	didr DidResolver

	signingKey *did.PrivKey
}

type DidResolver interface {
	GetDocument(ctx context.Context, didstr string) (*did.Document, error)
}

func NewKeyManager(didr plc.PLCClient, k *did.PrivKey) *KeyManager {
	return &KeyManager{
		didr:       didr,
		signingKey: k,
	}
}

type cachedKey struct {
	cachedAt time.Time
	pub      crypto.PublicKey
}

func (km *KeyManager) VerifyUserSignature(ctx context.Context, did string, sig []byte, msg []byte) error {
	ctx, span := otel.Tracer("keymgr").Start(ctx, "verifySignature")
	defer span.End()

	k, err := km.getKey(ctx, did)
	if err != nil {
		return err
	}

	return k.Verify(msg, sig)
}

func (km *KeyManager) getKey(ctx context.Context, did string) (*did.PubKey, error) {
	ctx, span := otel.Tracer("keymgr").Start(ctx, "getKey")
	defer span.End()

	// TODO: caching should be done at the DID document level, that way we can
	// have a thing that subscribes to plc updates for cache busting
	fmt.Println("GET KEY: ", did)
	doc, err := km.didr.GetDocument(ctx, did)
	if err != nil {
		return nil, err
	}

	pubk, err := doc.GetPublicKey("#signingKey")
	if err != nil {
		return nil, err
	}

	return pubk, nil
}

func (km *KeyManager) SignForUser(ctx context.Context, did string, msg []byte) ([]byte, error) {
	if km.signingKey == nil {
		return nil, fmt.Errorf("key manager does not have a signing key, cannot sign")
	}

	return km.signingKey.Sign(msg)
}
