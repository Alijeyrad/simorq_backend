package handler

import "github.com/gofiber/fiber/v3"

func ok(c fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{"data": data})
}

func created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": data})
}

func noContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func unauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

func forbidden(c fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
}

func notFound(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": msg})
}

func conflict(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": msg})
}

func tooManyRequests(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": msg})
}

func internalError(c fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
