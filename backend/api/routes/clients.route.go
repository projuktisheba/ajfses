package routes

import "github.com/go-chi/chi/v5"

// clientRoutes implements the routing for the ClientHandler.
func clientRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Client Routes ========

	// POST /: Create a new client (handles multipart form data with image)
	mux.Post("/", handlerRepo.Client.CreateClient)

	// GET /: Retrieve a list of all clients (using the base path for listing all/filtering)
	mux.Get("/", handlerRepo.Client.GetAllClients)

	// GET /: Retrieve client matrices
	mux.Get("/metrics", handlerRepo.Client.GetClientMetrics)

	// GET /{id}: Retrieve a single client by ID (ID is expected in the URL path)
	mux.Get("/profile/{id}", handlerRepo.Client.GetClient)

	// PUT /: Update an existing client (expects form data, uses query parameter {id} for identification)
	mux.Put("/", handlerRepo.Client.UpdateClient)

	// DELETE /: Delete a client (uses query parameter {id} for identification)
	mux.Delete("/", handlerRepo.Client.DeleteClient)

	return mux
}
