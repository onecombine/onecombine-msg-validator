package utils

import (
	"sync"

	"github.com/gofiber/fiber/v2"
)

var once sync.Once

type LoggingSession struct {
	values map[string]*Logger
}

var (
	instance LoggingSession
)

func GetLoggingSession() LoggingSession {

	once.Do(func() {
		instance.values = make(map[string]*Logger)
	})

	return instance
}

func (s *LoggingSession) Save(ctx *fiber.Ctx, logger *Logger) {
	id := ctx.Locals("Session-ID").(string)
	s.values[id] = logger
}

func (s *LoggingSession) Get(ctx *fiber.Ctx) *Logger {
	id := ctx.Locals("Session-ID").(string)
	logger := s.values[id]
	return logger
}

func (s *LoggingSession) Flush(ctx *fiber.Ctx) {
	id := ctx.Locals("Session-ID").(string)
	logger := s.values[id]
	logger.Print()
	delete(s.values, id)
}
