package ops

import (
	"context"
	"fmt"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

func (s *Service) ListProducts(ctx context.Context, status, q string, page, size int) (Page[Product], error) {
	page, size = clampPage(page, size)
	items, total, err := s.store.ListProducts(ctx, status, strings.TrimSpace(q), int32(size), int32((page-1)*size))
	if err != nil {
		return Page[Product]{}, err
	}
	return Page[Product]{Items: items, Total: total, Page: page, Size: size}, nil
}

func (s *Service) GetProduct(ctx context.Context, id int64) (Product, error) {
	return s.store.GetProduct(ctx, id)
}

func (s *Service) UpdateProduct(ctx context.Context, actor Actor, id int64, patch ProductPatch) (Product, error) {
	before, err := s.store.GetProduct(ctx, id)
	if err != nil {
		return Product{}, err
	}
	next := before
	locks := append([]string{}, before.LockedFields...)
	lock := func(field, old, neu string) {
		if neu != old && !contains(locks, field) {
			locks = append(locks, field)
		}
	}
	if patch.NameFa != nil {
		lock("name_fa", before.NameFa, *patch.NameFa)
		next.NameFa = strings.TrimSpace(*patch.NameFa)
	}
	if patch.NameEn != nil {
		lock("name_en", before.NameEn, *patch.NameEn)
		next.NameEn = strings.TrimSpace(*patch.NameEn)
	}
	if patch.GenericName != nil {
		lock("generic_name", before.GenericName, *patch.GenericName)
		next.GenericName = strings.TrimSpace(*patch.GenericName)
	}
	if patch.ImageURL != nil {
		lock("image_url", before.ImageURL, *patch.ImageURL)
		next.ImageURL = strings.TrimSpace(*patch.ImageURL)
	}
	if patch.BrandID != nil {
		next.BrandID = *patch.BrandID
	}
	if patch.CategoryID != nil {
		next.CategoryID = *patch.CategoryID
	}
	if patch.Status != nil {
		switch *patch.Status {
		case "published", "hidden", "needs_review":
			next.Status = *patch.Status
		default:
			return Product{}, fmt.Errorf("%w: invalid product status", ErrInvalidInput)
		}
	}
	next.LockedFields = locks
	if err := s.store.UpdateProduct(ctx, next); err != nil {
		return Product{}, err
	}
	got, err := s.store.GetProduct(ctx, id)
	if err != nil {
		return Product{}, err
	}
	s.audit(ctx, actor, "product", fmt.Sprint(id), "update", before, got)
	return got, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.store.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, actor Actor, in Category) (Category, error) {
	in.NameFa = strings.TrimSpace(in.NameFa)
	in.Slug = textfa.Slugify(in.Slug)
	if in.Slug == "" {
		in.Slug = textfa.Slugify(in.NameFa)
	}
	if in.NameFa == "" || in.Slug == "" {
		return Category{}, fmt.Errorf("%w: name and slug are required", ErrInvalidInput)
	}
	id, err := s.store.InsertCategory(ctx, in)
	if err != nil {
		return Category{}, err
	}
	got, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return Category{}, err
	}
	s.audit(ctx, actor, "category", fmt.Sprint(id), "create", nil, got)
	return got, nil
}

func (s *Service) UpdateCategory(ctx context.Context, actor Actor, id int64, in Category) (Category, error) {
	before, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return Category{}, err
	}
	in.ID = id
	in.NameFa = strings.TrimSpace(in.NameFa)
	in.Slug = textfa.Slugify(in.Slug)
	if in.NameFa == "" {
		in.NameFa = before.NameFa
	}
	if in.Slug == "" {
		in.Slug = before.Slug
	}
	if in.ParentID == id {
		return Category{}, fmt.Errorf("%w: category cannot be its own parent", ErrInvalidInput)
	}
	if err := s.store.UpdateCategory(ctx, in); err != nil {
		return Category{}, err
	}
	got, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return Category{}, err
	}
	s.audit(ctx, actor, "category", fmt.Sprint(id), "update", before, got)
	return got, nil
}

func (s *Service) DisableCategory(ctx context.Context, actor Actor, id, reassignTo int64) error {
	before, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return err
	}
	children, err := s.store.CountCategoryChildren(ctx, id)
	if err != nil {
		return err
	}
	if children > 0 {
		return ErrHasChildren
	}
	n, err := s.store.CountCategoryProducts(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 && reassignTo == 0 {
		return ErrHasProducts
	}
	if err := s.store.DeleteCategory(ctx, id, reassignTo); err != nil {
		return err
	}
	s.audit(ctx, actor, "category", fmt.Sprint(id), "disable", before, map[string]int64{"reassign_to": reassignTo})
	return nil
}

func (s *Service) ListBrands(ctx context.Context) ([]Brand, error) {
	return s.store.ListBrands(ctx)
}

func (s *Service) CreateBrand(ctx context.Context, actor Actor, in Brand) (Brand, error) {
	in.NameFa = strings.TrimSpace(in.NameFa)
	in.Slug = textfa.Slugify(in.Slug)
	if in.Slug == "" {
		in.Slug = textfa.Slugify(in.NameFa)
	}
	if in.NameFa == "" || in.Slug == "" {
		return Brand{}, fmt.Errorf("%w: name and slug are required", ErrInvalidInput)
	}
	id, err := s.store.InsertBrand(ctx, in)
	if err != nil {
		return Brand{}, err
	}
	got, err := s.store.GetBrand(ctx, id)
	if err != nil {
		return Brand{}, err
	}
	s.audit(ctx, actor, "brand", fmt.Sprint(id), "create", nil, got)
	return got, nil
}

func (s *Service) UpdateBrand(ctx context.Context, actor Actor, id int64, in Brand) (Brand, error) {
	before, err := s.store.GetBrand(ctx, id)
	if err != nil {
		return Brand{}, err
	}
	in.ID = id
	if strings.TrimSpace(in.NameFa) == "" {
		in.NameFa = before.NameFa
	}
	if strings.TrimSpace(in.Slug) == "" {
		in.Slug = before.Slug
	} else {
		in.Slug = textfa.Slugify(in.Slug)
	}
	if err := s.store.UpdateBrand(ctx, in); err != nil {
		return Brand{}, err
	}
	got, err := s.store.GetBrand(ctx, id)
	if err != nil {
		return Brand{}, err
	}
	s.audit(ctx, actor, "brand", fmt.Sprint(id), "update", before, got)
	return got, nil
}

func (s *Service) MergeBrands(ctx context.Context, actor Actor, fromID, intoID int64) error {
	if fromID == intoID {
		return ErrSameBrand
	}
	from, err := s.store.GetBrand(ctx, fromID)
	if err != nil {
		return err
	}
	into, err := s.store.GetBrand(ctx, intoID)
	if err != nil {
		return err
	}
	if err := s.store.MergeBrands(ctx, fromID, intoID); err != nil {
		return err
	}
	s.audit(ctx, actor, "brand", fmt.Sprint(fromID), "merge", from, into)
	return nil
}
