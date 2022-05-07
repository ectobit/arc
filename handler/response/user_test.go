package response_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler/response"
)

func TestFromDomainUser(t *testing.T) {
	t.Parallel()

	active := true
	now := time.Now()

	domainUser := &domain.User{
		ID:                 "926c7bed-18a7-4c0f-97fd-f5901b2c52ba",
		Email:              "john.doe@sixpack.com",
		Password:           []byte{},
		Activated:          nil,
		Created:            &now,
		Updated:            &now,
		ActivationToken:    "",
		PasswordResetToken: "",
		Active:             &active,
	}

	wantPublicUser := &response.User{
		ID:           domainUser.ID,
		Email:        domainUser.Email,
		Created:      &now,
		Updated:      &now,
		AuthToken:    "",
		RefreshToken: "",
	}

	gotPublicUser := response.FromDomainUser(domainUser)

	if diff := cmp.Diff(wantPublicUser, gotPublicUser); diff != "" {
		t.Errorf("FromDomainUser() mismatch (-want +got):\n%s", diff)
	}
}
