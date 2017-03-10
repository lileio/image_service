package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/lileio/image_service"
	"github.com/lileio/image_service/server"
	"github.com/lileio/image_service/storage"
	"github.com/lileio/image_service/workers"
	"github.com/lileio/lile"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("DEBUG") == "true" {
			logrus.SetLevel(logrus.DebugLevel)
		}

		store := storage.StorageFromEnv()

		poolSize := 5
		if os.Getenv("WORKER_POOL_SIZE") != "" {
			i, err := strconv.Atoi(os.Getenv("WORKER_POOL_SIZE"))
			if err != nil {
				panic(err)
			}

			poolSize = i
		}

		workers.StartWorkerPool(poolSize, store)
		s := &server.Server{}

		impl := func(g *grpc.Server) {
			image_service.RegisterImageServiceServer(g, s)
		}

		err := lile.NewServer(
			lile.Name("image_service"),
			lile.Implementation(impl),
		).ListenAndServe()

		log.Fatal(err)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
