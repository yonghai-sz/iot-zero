// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"
	"errors"
	"sort"

	"iot-zero/services/platform-api/internal/session"
	"iot-zero/services/platform-api/internal/svc"
	"iot-zero/services/platform-api/internal/types"
	"iot-zero/services/platform-api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuTreeLogic) GetMenuTree() (*types.GetMenuTreeResp, error) {
	roleType, ok := session.RoleTypeFromContext(l.ctx)
	if !ok {
		return nil, errUnauthorized
	}

	role, err := l.svcCtx.RoleModel.FindOneByRoleType(l.ctx, roleType)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errRoleNotFound
		}
		return nil, err
	}

	menuIds, err := l.svcCtx.RoleMenuModel.FindMenuIdsByRoleId(l.ctx, role.Id)
	if err != nil {
		return nil, err
	}

	menus, err := l.svcCtx.MenuModel.FindByIds(l.ctx, menuIds)
	if err != nil {
		return nil, err
	}

	allMenus, err := l.collectWithAncestors(menus)
	if err != nil {
		return nil, err
	}

	return &types.GetMenuTreeResp{
		List: buildMenuTree(allMenus),
	}, nil
}

func (l *GetMenuTreeLogic) collectWithAncestors(menus []*model.Menu) ([]*model.Menu, error) {

	byId := make(map[uint64]*model.Menu, len(menus))
	for _, menu := range menus {
		byId[menu.Id] = menu
	}

	unresolved := make(map[uint64]struct{})
	pending := uniqueMissingParentIds(byId, unresolved)

	for len(pending) > 0 {
		parents, err := l.svcCtx.MenuModel.FindByIds(l.ctx, pending)
		if err != nil {
			return nil, err
		}
		found := make(map[uint64]struct{}, len(parents))
		for _, parent := range parents {
			byId[parent.Id] = parent
			found[parent.Id] = struct{}{}
		}
		for _, id := range pending {
			if _, ok := found[id]; !ok {
				unresolved[id] = struct{}{}
			}
		}
		pending = uniqueMissingParentIds(byId, unresolved)
	}

	result := make([]*model.Menu, 0, len(byId))
	for _, menu := range byId {
		result = append(result, menu)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Id < result[j].Id })

	return result, nil
}

func uniqueMissingParentIds(byId map[uint64]*model.Menu, unresolved map[uint64]struct{}) []uint64 {

	seen := make(map[uint64]struct{})

	var ids []uint64

	for _, menu := range byId {
		if menu.ParentId == 0 {
			continue
		}
		if _, exists := byId[menu.ParentId]; exists {
			continue
		}
		if _, skip := unresolved[menu.ParentId]; skip {
			continue
		}
		if _, dup := seen[menu.ParentId]; dup {
			continue
		}
		seen[menu.ParentId] = struct{}{}
		ids = append(ids, menu.ParentId)
	}
	return ids
}

func buildMenuTree(menus []*model.Menu) []types.MenuInfo {
	children := make(map[uint64][]*model.Menu)
	var roots []*model.Menu
	for _, menu := range menus {
		if menu.ParentId == 0 {
			roots = append(roots, menu)
			continue
		}
		children[menu.ParentId] = append(children[menu.ParentId], menu)
	}

	tree := make([]types.MenuInfo, 0, len(roots))
	for _, root := range roots {
		tree = append(tree, attachMenuChildren(root, children))
	}
	return tree
}

func attachMenuChildren(menu *model.Menu, children map[uint64][]*model.Menu) types.MenuInfo {
	node := toMenuInfo(menu)
	for _, child := range children[menu.Id] {
		node.Children = append(node.Children, attachMenuChildren(child, children))
	}
	return node
}
