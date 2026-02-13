package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wyfcoding/ecommerce/internal/userprofile/application"
	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
)

type UserProfileHandler struct {
	commandService *application.ProfileCommandService
	queryService   *application.ProfileQueryService
}

func NewUserProfileHandler(
	commandService *application.ProfileCommandService,
	queryService *application.ProfileQueryService,
) *UserProfileHandler {
	return &UserProfileHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

func (h *UserProfileHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/user-profiles/{user_id}", h.GetProfile).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles", h.CreateProfile).Methods("POST")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/tags", h.GetTags).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/tags", h.AddTag).Methods("POST")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/tags/{tag_key}", h.RemoveTag).Methods("DELETE")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/behaviors", h.RecordBehavior).Methods("POST")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/behavior-features", h.GetBehaviorFeatures).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/preferences", h.GetPreferences).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/consumption", h.GetConsumptionProfile).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/recalculate", h.RecalculateProfile).Methods("POST")
	router.HandleFunc("/api/v1/user-profiles/{user_id}/summary", h.GetProfileSummary).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/by-tag", h.GetUsersByTag).Methods("GET")
	router.HandleFunc("/api/v1/user-profiles/by-segment/{segment_no}", h.GetUsersBySegment).Methods("GET")
}

func (h *UserProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint64 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	profile, err := h.commandService.CreateProfile(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func (h *UserProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	profile, err := h.queryService.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func (h *UserProfileHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	tags, err := h.queryService.GetTags(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tags)
}

func (h *UserProfileHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req struct {
		TagKey     string  `json:"tag_key"`
		TagValue   string  `json:"tag_value"`
		Category   int8    `json:"category"`
		Source     int8    `json:"source"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.commandService.AddTag(r.Context(), userID, req.TagKey, req.TagValue, domain.TagCategory(req.Category), domain.TagSource(req.Source), req.Confidence); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	tagKey := mux.Vars(r)["tag_key"]

	if err := h.commandService.RemoveTag(r.Context(), userID, tagKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) RecordBehavior(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req struct {
		BehaviorType string `json:"behavior_type"`
		TargetType   string `json:"target_type"`
		TargetID     uint64 `json:"target_id"`
		Value        string `json:"value"`
		Duration     int64  `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.commandService.RecordBehavior(r.Context(), userID, domain.BehaviorType(req.BehaviorType), req.TargetType, req.TargetID, req.Value, req.Duration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) GetBehaviorFeatures(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	features, err := h.queryService.GetBehaviorFeatures(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(features)
}

func (h *UserProfileHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	preferences, err := h.queryService.GetPreferences(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(preferences)
}

func (h *UserProfileHandler) GetConsumptionProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	consumption, err := h.queryService.GetConsumptionProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(consumption)
}

func (h *UserProfileHandler) RecalculateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.commandService.RecalculateProfile(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *UserProfileHandler) GetProfileSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	summary, err := h.queryService.GetProfileSummary(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(summary)
}

func (h *UserProfileHandler) GetUsersByTag(w http.ResponseWriter, r *http.Request) {
	tagKey := r.URL.Query().Get("tag_key")
	tagValue := r.URL.Query().Get("tag_value")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	profiles, err := h.queryService.GetUsersByTag(r.Context(), tagKey, tagValue, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profiles)
}

func (h *UserProfileHandler) GetUsersBySegment(w http.ResponseWriter, r *http.Request) {
	segmentNo := mux.Vars(r)["segment_no"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	profiles, err := h.queryService.GetSegmentUsers(r.Context(), segmentNo, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profiles)
}
