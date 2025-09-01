package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type numberRequest struct {
	Value int `json:"value"`
}

func main() {
	// Create the shared store (concurrency-safe via mutex).
	store := NewStore()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Read current list
	r.GET("/numbers", func(c *gin.Context) {
		state := store.Snapshot()
		c.JSON(http.StatusOK, gin.H{
			"list": state,
		})
	})

	// Apply a number
	r.POST("/numbers", func(c *gin.Context) {
		var req numberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON. Expected {\"value\": <int>}"})
			return
		}
		action, updated := store.Apply(req.Value)
		c.JSON(http.StatusOK, gin.H{
			"input":   req.Value,
			"action":  action,
			"updated": updated,
		})
	})

	// Reset list (useful for testing/demo)
	r.POST("/reset", func(c *gin.Context) {
		store.Reset()
		c.JSON(http.StatusOK, gin.H{"message": "state cleared"})
	})

	// Run the given example sequence: 5, 10, -6
	r.POST("/example", func(c *gin.Context) {
		store.Reset()
		type step struct {
			Input   int   `json:"input"`
			List    []int `json:"list"`
			Action  string`json:"action"`
		}
		seq := []int{5, 10, -6}
		var steps []step

		for _, v := range seq {
			action, after := store.Apply(v)
			steps = append(steps, step{Input: v, List: append([]int(nil), after...), Action: action})
		}
		c.JSON(http.StatusOK, gin.H{
			"sequence": seq,
			"steps":    steps,
			"final":    store.Snapshot(),
		})
	})

	addr := ":8080"
	log.Printf("🚀 starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
