---
title: Recipe Search & Meal Planning
docpush:
  confluence:
    space: COOKBOOK
    pageId: 123456789
---

# Recipe Search & Meal Planning

## Summary

The Recipe Search feature lets users discover recipes by ingredient, cuisine, or dietary restriction, then organize them into weekly meal plans with automatic grocery list generation.

## Architecture

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Frontend | React 19 | Search UI + meal planner |
| API | Go + Chi | REST endpoints |
| Search | Elasticsearch | Full-text recipe search |
| Storage | PostgreSQL | Recipe and plan data |

## Key Features

- **Smart Search**: Full-text search with filters for cuisine, prep time, and dietary tags
- **Meal Planner**: Drag-and-drop weekly calendar for scheduling meals
- **Grocery Lists**: Auto-generated shopping lists from planned recipes
- **Favorites**: Save and organize recipes into custom collections

## API Endpoints

```go
// Search recipes
GET /api/recipes?q=pasta&cuisine=italian&maxTime=30

// Get recipe by ID
GET /api/recipes/:id

// Create a meal plan
POST /api/plans
```

## Database Schema

```sql
CREATE TABLE recipes (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  cuisine VARCHAR(50),
  prep_time_minutes INT,
  servings INT DEFAULT 4,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);
```

> **Note**: All meal plans are scoped to a `user_id` for data isolation.

1. Search for recipes
2. Add to meal plan
3. Adjust servings
4. Generate grocery list
5. Mark items as purchased

## Links

- [RFC](https://example.com/rfc)
- [Design Doc](https://example.com/design)
