package api

import (
	"context"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/types"
	"net/http"
	"strconv"
	"strings"
)

type ServerConfig struct {
	Logger log.Logger
	Addr   string
}

type Server struct {
	ServerConfig
	chain *core.Blockchain
	srv   *http.Server
}

func NewServer(cfg ServerConfig, chain *core.Blockchain) *Server {
	return &Server{
		ServerConfig: cfg,
		chain:        chain,
	}
}

func (s *Server) Start() error {
	eg := s.SetRouter()
	srv := &http.Server{
		Addr:    s.Addr,
		Handler: eg,
	}
	s.srv = srv
	return srv.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.srv.Shutdown(context.Background())
}

func (s *Server) SetRouter() *gin.Engine {
	eg := gin.Default()
	eg.GET("/block/:hash", s.handleGetBlock)
	eg.GET("/tx/:hash", s.handleGetTransaction)
	eg.GET("/test", s.handleTest)
	return eg
}

func (s *Server) handleGetBlock(ctx *gin.Context) {
	heightOrHash := ctx.Param("hash")
	if height, err := strconv.ParseUint(heightOrHash, 10, 64); err == nil {
		block, err := s.chain.GetBlock(height)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"msg":   "failed to get block",
				"error": err,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"msg":   "success",
			"block": block,
		})
		return
	}
	if strings.HasPrefix(heightOrHash, "0x") {
		heightOrHash = heightOrHash[2:]
	}
	hashByte, err := hex.DecodeString(heightOrHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "failed to decode hash",
			"error": err,
		})
		return
	}
	block, err := s.chain.GetBlockByHash(types.HashFromBytes(hashByte))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "failed to get block",
			"error": err,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":   "success",
		"block": block,
	})

}

func (s *Server) handleTest(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"msg": "success"})
}

func (s *Server) handleGetTransaction(ctx *gin.Context) {
	heightOrHash := ctx.Param("hash")
	if height, err := strconv.ParseUint(heightOrHash, 10, 64); err == nil {
		block, err := s.chain.GetBlock(height)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"msg":   "failed to get block",
				"error": err,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"msg":          "success",
			"transactions": block.Transactions,
		})
		return
	}
	if strings.HasPrefix(heightOrHash, "0x") {
		heightOrHash = heightOrHash[2:]
	}
	hashByte, err := hex.DecodeString(heightOrHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "failed to decode hash",
			"error": err,
		})
		return
	}
	transactions, err := s.chain.GetTransactionByHash(types.HashFromBytes(hashByte))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":   "failed to get block",
			"error": err,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":          "success",
		"transactions": transactions,
	})
}
