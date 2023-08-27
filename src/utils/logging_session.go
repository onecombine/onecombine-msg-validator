package utils

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var once sync.Once

type LoggingSession struct {
	store *session.Store
}

var (
	instance LoggingSession
)

func GetLoggingSession() LoggingSession {

	once.Do(func() {
		instance.store = session.New()
	})

	return instance
}

func (s *LoggingSession) Save(ctx *fiber.Ctx, logger *Logger) {
	sess, err := s.store.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sess.Set("logger", logger)
	sess.Save()
}

func (s *LoggingSession) Get(ctx *fiber.Ctx) *Logger {
	sess, err := s.store.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}

	logger := sess.Get("logger").(*Logger)
	return logger
}

func (s *LoggingSession) Flush(ctx *fiber.Ctx) {
	sess, err := s.store.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	logger := sess.Get("logger").(*Logger)
	logger.Print()
	sess.Destroy()
}
