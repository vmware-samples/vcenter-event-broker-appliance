package function

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

// vsClient is a client for vSphere.
type vsClient struct {
	govmomi *govmomi.Client
	rest    *rest.Client
}

func newClient(ctx context.Context, u url.URL, insecure bool) (*vsClient, error) {
	var clt vsClient

	gc, err := govmomi.NewClient(ctx, &u, insecure)
	if err != nil {
		return nil, fmt.Errorf("connecting to govmomi api failed: %w", err)
	}
	clt.govmomi = gc

	clt.rest = rest.NewClient(clt.govmomi.Client)
	err = clt.rest.Login(ctx, u.User)
	if err != nil {
		return nil, fmt.Errorf("log in to rest api failed: %w", err)
	}

	return &clt, nil
}

// moTag adds an existing tag to a VirtualMachine.
func (clt *vsClient) moTag(ctx context.Context, vm types.ManagedObjectReference, tagID string) error {
	// Get the tag manager which does the tagging.
	m := tags.NewManager(clt.rest)

	// Attach tag to VM.
	err := m.AttachTag(ctx, tagID, vm)
	if err != nil {
		return fmt.Errorf("attach tag to VM failed: %w", err)
	}

	return nil
}

func (clt *vsClient) logout(ctx context.Context) error {
	err := clt.govmomi.Logout(ctx)
	if err != nil {
		return fmt.Errorf("govmomi api logout failed: %w", err)
	}

	err = clt.rest.Logout(ctx)
	if err != nil {
		return fmt.Errorf("rest api logout failed: %w", err)
	}

	return nil
}
