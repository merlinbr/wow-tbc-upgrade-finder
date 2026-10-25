package server

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

var errInvalidQuery = errors.New("invalid query")

type itemRow struct {
	domain.Item
	Compatible bool   `json:"compatible"`
	Reason     string `json:"reason,omitempty"`
}

type itemDetail struct {
	domain.Item
	Compatible       bool   `json:"compatible"`
	Reason           string `json:"reason,omitempty"`
	LevelRequirement string `json:"levelRequirement"`
}

func (a *api) handleItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	slot := domain.Slot(query.Get("slot"))
	if slot != "" && !catalog.ValidSlot(slot) {
		writeError(w, http.StatusBadRequest, "invalid_query", "unknown slot")
		return
	}
	limit, offset, err := parsePagination(query.Get("limit"), query.Get("offset"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}

	filter := catalog.ItemFilter{
		Class:  query.Get("class"),
		Slot:   slot,
		Query:  query.Get("q"),
		Source: query.Get("source"),
		Limit:  limit,
		Offset: offset,
	}
	items, total := catalog.SearchItems(a.catalog, filter)

	rows := make([]itemRow, 0, len(items))
	for index := range items {
		row := itemRow{Item: items[index], Compatible: true}
		if filter.Class != "" {
			row.Compatible, row.Reason = a.catalog.Compatibility(
				domain.Character{Class: filter.Class}, &items[index], filter.Slot,
			)
		}
		rows = append(rows, row)
	}

	writeJSON(w, http.StatusOK, struct {
		Items  []itemRow `json:"items"`
		Total  int       `json:"total"`
		Limit  int       `json:"limit"`
		Offset int       `json:"offset"`
	}{Items: rows, Total: total, Limit: effectiveLimit(limit), Offset: max(offset, 0)})
}

func (a *api) handleItemDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, "not_found", "item not found")
		return
	}
	item, ok := a.catalog.Items[id]
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "item not found")
		return
	}
	query := r.URL.Query()
	slot := domain.Slot(query.Get("slot"))
	if slot != "" && !catalog.ValidSlot(slot) {
		writeError(w, http.StatusBadRequest, "invalid_query", "unknown slot")
		return
	}
	class := domain.Class(query.Get("class"))
	compatible := true
	reason := ""
	if class != "" {
		compatible, reason = a.catalog.Compatibility(
			domain.Character{Class: class}, item, slot,
		)
	} else if slot != "" {
		compatible = slices.Contains(item.Slots, slot)
		if !compatible {
			reason = "incompatible_slot"
		}
	}

	writeJSON(w, http.StatusOK, itemDetail{
		Item:             *item,
		Compatible:       compatible,
		Reason:           reason,
		LevelRequirement: "unknown",
	})
}

func (a *api) handleEnchants(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	slot := domain.Slot(query.Get("slot"))
	if slot != "" && !catalog.ValidSlot(slot) {
		writeError(w, http.StatusBadRequest, "invalid_query", "unknown slot")
		return
	}
	limit, offset, err := parsePagination(query.Get("limit"), query.Get("offset"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	enchants := catalog.SearchEnchants(a.catalog, catalog.EnchantFilter{
		Class:  query.Get("class"),
		Slot:   slot,
		Query:  query.Get("q"),
		Limit:  limit,
		Offset: offset,
	})
	writeJSON(w, http.StatusOK, struct {
		Enchants []domain.Enchant `json:"enchants"`
		Count    int              `json:"count"`
	}{Enchants: enchants, Count: len(enchants)})
}

func (a *api) handleTalents(w http.ResponseWriter, r *http.Request) {
	class := domain.Class(r.URL.Query().Get("class"))
	trees, err := catalog.TalentTrees(a.catalog, class)
	if err != nil {
		if errors.Is(err, domain.ErrUnsupportedClass) {
			writeError(w, http.StatusBadRequest, "unsupported_class", "unknown class")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "talent lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, trees)
}

func parsePagination(limitRaw, offsetRaw string) (int, int, error) {
	limit := 0
	offset := 0
	if limitRaw != "" {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("limit must be a non-negative integer")
		}
		limit = parsed
	}
	if offsetRaw != "" {
		parsed, err := strconv.Atoi(offsetRaw)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("offset must be a non-negative integer")
		}
		offset = parsed
	}
	return limit, offset, nil
}

func effectiveLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}
