package user

import (
	"context"
	"errors"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/model"
)

func findUserByUsername(ctx context.Context, userModel model.UserModel, username string) (*model.User, error) {
	entity, err := userModel.FindOneByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errUserNotFound
		}
		return nil, err
	}
	return entity, nil
}

func loadUserForAdmin(ctx context.Context, userModel model.UserModel, username string) (*model.User, error) {
	if err := authz.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	entity, err := findUserByUsername(ctx, userModel, username)
	if err != nil {
		return nil, err
	}

	if err = authz.AuthorizeAdminForTenant(ctx, entity.TenantId); err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return nil, errUserNotFound
		}
		return nil, err
	}
	return entity, nil
}

func loadUserForSelfOrAdmin(ctx context.Context, userModel model.UserModel, username string) (*model.User, error) {
	if err := authz.AuthorizeSelfOrAdmin(ctx, username); err != nil {
		return nil, err
	}

	entity, err := findUserByUsername(ctx, userModel, username)
	if err != nil {
		return nil, err
	}

	if operatorUsername, ok := session.UsernameFromContext(ctx); ok && operatorUsername == username {
		return entity, nil
	}

	if err = authz.AuthorizeAdminForTenant(ctx, entity.TenantId); err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return nil, errUserNotFound
		}
		return nil, err
	}
	return entity, nil
}
