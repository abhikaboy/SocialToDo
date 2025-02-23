package task

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/abhikaboy/SocialToDo/internal/xvalidator"
	"github.com/abhikaboy/SocialToDo/xutils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var validator = xvalidator.Validator
type Handler struct {
	service *Service
}

func (h *Handler) GetTasksByUser(c *fiber.Ctx) error {
	user_id := c.UserContext().Value("user_id").(string)
	slog.LogAttrs(c.Context(), slog.LevelInfo, "User ID", slog.String("user_id", user_id))

	id := c.Query("id", user_id) // uses the logged in user if not specified
	userId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format " + err.Error() + " ID: " + id,
		})
	}

	var sort SortParams

	sort.SortBy = c.Query("sortBy", "timestamp")
	sort.SortDir, err = strconv.Atoi(c.Query("sortDir", "-1"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid sortDir format",
		})
	}

  sortAggregation := bson.D{
		{Key: "$sort", Value: bson.M{
			sort.SortBy: sort.SortDir,
		}},
	}

	Tasks, err := h.service.GetTasksByUser(userId, sortAggregation)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	return c.JSON(Tasks)
}

func (h *Handler) CreateTask(c *fiber.Ctx) error {
	var params CreateTaskParams

	user_id := c.UserContext().Value("user_id").(string)

	err, ids := xutils.ParseIDs(c, c.Params("category"), user_id)
	if err != nil {
		slog.LogAttrs(c.Context(), slog.LevelError, "Error Parsing IDs", slog.String("error", err.Error()))
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	userId, categoryId := ids[1], ids[0]

	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validator.Validate(params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	doc := TaskDocument{
		ID:        primitive.NewObjectID(),
		Priority:  params.Priority,
		Content:   params.Content,
		Value:     params.Value,
		Recurring: params.Recurring,
		RecurDetails: params.RecurDetails,
		Public:    params.Public,
		Active:    params.Active,
		Timestamp: time.Now(),
	}

	_, err = h.service.CreateTask(userId, categoryId, &doc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	return c.Status(fiber.StatusCreated).JSON(doc)
}

func (h *Handler) GetTasks(c *fiber.Ctx) error {
	Tasks, err := h.service.GetAllTasks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch Tasks",
		})
	}

	return c.JSON(Tasks)
}

func (h *Handler) GetTask(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	Task, err := h.service.GetTaskByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Task not found",
		})
	}

	return c.JSON(Task)
}

func (h *Handler) UpdateTask(c *fiber.Ctx) error {
	context_id := c.UserContext().Value("user_id").(string)

	user_id, err := primitive.ObjectIDFromHex(context_id)
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	categoryId, err := primitive.ObjectIDFromHex(c.Params("category"))
	
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	var update UpdateTaskDocument
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if _, err := h.service.UpdatePartialTask(user_id, id, categoryId, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	if err := h.service.DeleteTask(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete Task",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
