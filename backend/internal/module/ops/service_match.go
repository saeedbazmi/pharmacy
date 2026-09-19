package ops

import (
	"context"
	"fmt"
)

func (s *Service) ListMatches(ctx context.Context, sourceID int64, minScore float64, page, size int) (Page[Match], error) {
	page, size = clampPage(page, size)
	items, total, err := s.store.ListMatches(ctx, sourceID, minScore, int32(size), int32((page-1)*size))
	if err != nil {
		return Page[Match]{}, err
	}
	return Page[Match]{Items: items, Total: total, Page: page, Size: size}, nil
}

func (s *Service) ApproveMatch(ctx context.Context, actor Actor, id int64) error {
	row, err := s.store.GetMatch(ctx, id)
	if err != nil {
		return err
	}
	if row.Status != "pending" {
		return ErrMatchNotPending
	}
	productID := row.SuggestedProductID
	if productID == 0 {
		return fmt.Errorf("%w: no suggested product", ErrInvalidInput)
	}
	if err := s.store.ApplyMatch(ctx, applyMatchInput{
		ID: id, SourceItemID: row.SourceItemID, PharmacyID: row.PharmacyID,
		ProductID: productID, PreviousProductID: row.CurrentProductID,
		ActorID: actor.ID, Status: "approved",
	}); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "match.approved", "actor_id", actor.ID, "match_id", id, "product_id", productID)
	s.audit(ctx, actor, "match_candidate", fmt.Sprint(id), "approve", row.Match, map[string]int64{"product_id": productID})
	return nil
}

func (s *Service) RejectMatch(ctx context.Context, actor Actor, id int64) error {
	row, err := s.store.GetMatch(ctx, id)
	if err != nil {
		return err
	}
	if row.Status != "pending" {
		return ErrMatchNotPending
	}
	if err := s.store.ApplyMatch(ctx, applyMatchInput{
		ID: id, SourceItemID: row.SourceItemID, PharmacyID: row.PharmacyID,
		ProductID: row.SuggestedProductID, PreviousProductID: row.CurrentProductID,
		ActorID: actor.ID, Status: "rejected",
	}); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "match.rejected", "actor_id", actor.ID, "match_id", id)
	s.audit(ctx, actor, "match_candidate", fmt.Sprint(id), "reject", row.Match, nil)
	return nil
}

func (s *Service) LinkMatch(ctx context.Context, actor Actor, id, productID int64) error {
	if productID <= 0 {
		return fmt.Errorf("%w: product_id is required", ErrInvalidInput)
	}
	if _, err := s.store.GetProduct(ctx, productID); err != nil {
		return err
	}
	row, err := s.store.GetMatch(ctx, id)
	if err != nil {
		return err
	}
	if row.Status != "pending" {
		return ErrMatchNotPending
	}
	if err := s.store.ApplyMatch(ctx, applyMatchInput{
		ID: id, SourceItemID: row.SourceItemID, PharmacyID: row.PharmacyID,
		ProductID: productID, PreviousProductID: row.CurrentProductID,
		ActorID: actor.ID, Status: "approved",
	}); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "match.approved", "actor_id", actor.ID, "match_id", id, "product_id", productID, "manual", true)
	s.audit(ctx, actor, "match_candidate", fmt.Sprint(id), "link", row.Match, map[string]int64{"product_id": productID})
	return nil
}

func (s *Service) CreateFromMatch(ctx context.Context, actor Actor, id int64) (int64, error) {
	row, err := s.store.GetMatch(ctx, id)
	if err != nil {
		return 0, err
	}
	if row.Status != "pending" {
		return 0, ErrMatchNotPending
	}
	productID, err := s.store.CreateProductFromItem(ctx, row, actor.ID)
	if err != nil {
		return 0, err
	}
	s.log.InfoContext(ctx, "match.approved", "actor_id", actor.ID, "match_id", id, "product_id", productID, "created", true)
	s.audit(ctx, actor, "match_candidate", fmt.Sprint(id), "create_product", row.Match, map[string]int64{"product_id": productID})
	return productID, nil
}

func (s *Service) UndoMatch(ctx context.Context, actor Actor, id int64) error {
	row, err := s.store.GetMatch(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == "pending" {
		return ErrMatchNotDecided
	}
	if err := s.store.ReopenMatch(ctx, id, row.PreviousProductID); err != nil {
		return err
	}
	s.audit(ctx, actor, "match_candidate", fmt.Sprint(id), "undo", row.Match, map[string]int64{"previous_product_id": row.PreviousProductID})
	return nil
}
