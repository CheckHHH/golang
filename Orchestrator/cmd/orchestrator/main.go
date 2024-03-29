package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/CheckHHH/golang/Orchestrator/internal/config"
	"github.com/CheckHHH/golang/Orchestrator/internal/http/server"
	"github.com/CheckHHH/golang/Orchestrator/internal/storage/memory"
	"github.com/CheckHHH/golang/Orchestrator/internal/storage/postgresql/postgresql_ast"
	"github.com/CheckHHH/golang/Orchestrator/internal/storage/postgresql/postgresql_config"
	"github.com/CheckHHH/golang/Orchestrator/internal/tasks/queue"
	serverTCP "github.com/CheckHHH/golang/Orchestrator/internal/tcp/server"
)

func main() {
	conf := config.New()
	postgresql_config.Load(conf)
	storage := memory.New(conf)
	newQueue := queue.NewMapQueue(queue.NewLockFreeQueue(), conf)
	postgresql_ast.GetAll(conf, newQueue, storage)
	postgresql_ast.Update(conf, newQueue, storage)
	tcpServer, err := serverTCP.NewServer(":"+conf.TCPPort, conf, newQueue, storage)
	if err != nil {
		slog.Error("Ошибка запуска TCP/IP сервера:", "ошибка:", err)
		os.Exit(1)
	}
	slog.Info("Запуск TCP/IP сервера на порту " + conf.TCPPort)
	tcpServer.Start()
	slog.Info("Оркестратор запущен")
	ctx, cancel := context.WithCancel(context.Background())
	shutDown, err := server.Run(ctx, conf, newQueue, storage)
	if err != nil {
		slog.Error("Ошибка запуска сервера:", "ошибка:", err)
		os.Exit(1)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-c
	cancel()
	tcpServer.Stop()
	shutDown(ctx)
	slog.Info("Сервер TCP/IP остановлен")
	slog.Info("Оркестратор остановлен")
	os.Exit(0)
}
