package utils

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var once sync.Once

type LoggingSession struct {
	store  *session.Store
	values map[string]*Logger
}

var (
	instance LoggingSession
)

func GetLoggingSession() LoggingSession {

	once.Do(func() {
		instance.store = session.New()
		instance.values = make(map[string]*Logger)
	})

	return instance
}

func (s *LoggingSession) Save(ctx *fiber.Ctx, logger *Logger) {
	sess, err := s.store.Get(ctx)
	if err != nil {
		panic(err)
	}
	s.values[sess.ID()] = logger
}

func (s *LoggingSession) Get(ctx *fiber.Ctx) *Logger {
	sess, err := s.store.Get(ctx)
	if err != nil {
		panic(err)
	}

	logger := s.values[sess.ID()]
	return logger
}

func (s *LoggingSession) Flush(ctx *fiber.Ctx) {
	sess, err := s.store.Get(ctx)
	if err != nil {
		panic(err)
	}
	logger := s.values[sess.ID()]
	logger.Print()
	delete(s.values, sess.ID())
	sess.Destroy()
}
