# Database Layer Documentation

This document covers the Saga API's database abstraction layer, including interfaces, transaction patterns, and SurrealDB-specific handling.

## Table of Contents

- [Database Interface](#database-interface)
- [Transaction Patterns](#transaction-patterns)
- [SurrealDB Specifics](#surrealdb-specifics)
- [Repository Pattern](#repository-pattern)
- [Error Handling](#error-handling)
- [Code Examples](#code-examples)

---

## Database Interface

**Location:** `internal/database/database.go`

The `Database` interface abstracts away SurrealDB specifics, allowing for potential future database swaps and easier testing.

### Interface Definition

```go
type Database interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error

    // Query execution
    Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error)
    QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error)
    Execute(ctx context.Context, query string, vars map[string]interface{}) error

    // Transaction support
    BeginTx(ctx context.Context) (Transaction, error)
}
```

### Method Comparison

| Method | Returns | Use Case |
|--------|---------|----------|
| `Query` | `[]interface{}` | SELECT queries returning multiple rows |
| `QueryOne` | `interface{}` | SELECT queries expecting single row |
| `Execute` | `error` | CREATE/UPDATE/DELETE without result needs |

### Error Types

```go
var (
    ErrNotFound      = errors.New("record not found")
    ErrDuplicate     = errors.New("duplicate record")
    ErrConnection    = errors.New("database connection error")
    ErrQuery         = errors.New("query execution error")
    ErrLimitExceeded = errors.New("result limit exceeded")
)
```

---

## Transaction Patterns

**CRITICAL:** Saga uses **batch-based transactions**, NOT connection-level isolation.

### Understanding Batch-Based Transactions

Unlike traditional databases where `BeginTx()` creates an isolated connection:

```
Traditional (PostgreSQL):              Saga (SurrealDB):
┌──────────────────────────┐          ┌──────────────────────────┐
│  BEGIN                   │          │  Queries accumulated     │
│  INSERT INTO users...    │  ←→      │  in memory until         │
│  UPDATE accounts...      │          │  Commit() is called      │
│  COMMIT                  │          │                          │
│                          │          │  Then wrapped in:        │
│  (Connection isolated    │          │  BEGIN TRANSACTION;      │
│   during entire block)   │          │  query1;                 │
│                          │          │  query2;                 │
│                          │          │  COMMIT TRANSACTION;     │
└──────────────────────────┘          └──────────────────────────┘
```

**Implications:**
1. No isolation between `Add()` calls until `Commit()`
2. All queries execute together atomically at commit time
3. Rollback before commit simply discards accumulated queries

### Pattern 1: AtomicBatch (Recommended)

**Location:** `internal/database/transaction.go`

The simplest and most common pattern for multi-statement atomic operations.

```go
// Example: Award resonance points atomically
func (r *ResonanceRepository) AwardPointsAtomic(ctx context.Context, entry *model.ResonanceLedgerEntry) error {
    batch := database.NewAtomicBatch()

    // 1. Create ledger entry
    batch.Add(`
        CREATE resonance_ledger CONTENT {
            user: $user,
            stat: $stat,
            points: $points,
            source_object_id: $source_id,
            reason: $reason,
            created_on: time::now()
        }
    `, map[string]interface{}{
        "user":      entry.User,
        "stat":      entry.Stat,
        "points":    entry.Points,
        "source_id": entry.SourceObjectID,
        "reason":    entry.Reason,
    })

    // 2. Update daily cap
    batch.Add(`
        UPSERT resonance_daily_cap SET
            user = $user,
            date = time::format(time::now(), "%Y-%m-%d"),
            points_earned += $points,
            updated_on = time::now()
        WHERE user = $user AND date = time::format(time::now(), "%Y-%m-%d")
    `, map[string]interface{}{
        "user":   entry.User,
        "points": entry.Points,
    })

    // 3. Trigger score recalculation (handled by DB trigger)
    // No additional query needed - the resonance_score_update event fires

    return batch.Execute(ctx, r.db)  // All or nothing
}
```

**When to use:** 2-5 statements that must succeed together.

### Pattern 2: TxBuilder (Variable Namespacing)

**Location:** `internal/database/transaction.go`

Use when you need to combine queries with potentially conflicting variable names.

```go
func CreateUserWithProfile(ctx context.Context, db database.Database, user *model.User, profile *model.UserProfile) error {
    tb := database.NewTxBuilder()

    // Add first query - variables get namespaced automatically
    // $email becomes $1_email
    tb.Add(`
        CREATE user CONTENT {
            email: $email,
            username: $username,
            created_on: time::now()
        }
    `, map[string]interface{}{
        "email":    user.Email,
        "username": user.Username,
    })

    // Add second query - $email here becomes $2_email (no collision)
    tb.Add(`
        CREATE user_profile CONTENT {
            user: $user_id,
            bio: $bio,
            discovery_eligible: false,
            created_on: time::now()
        }
    `, map[string]interface{}{
        "user_id": "user:" + user.ID,
        "bio":     profile.Bio,
    })

    // Execute atomically
    _, err := database.ExecuteTransaction(ctx, db, tb)
    return err
}
```

**How namespacing works:**
```
Original:  $email, $username, $bio
After:     $v1_email, $v1_username, $v2_bio
```

Note the `v` prefix before the number - this is required by SurrealDB variable naming rules. This prevents collisions when combining queries from different sources.

### Pattern 3: UnitOfWork (Service Layer)

**Location:** `internal/database/transaction.go`

Use when you need custom rollback logic for failed operations.

```go
func (s *EventService) CreateEventWithNotifications(ctx context.Context, event *model.Event) error {
    uow := database.NewUnitOfWork(s.db)

    // Add event creation
    uow.Add(`CREATE event CONTENT {...}`, eventVars)

    // Add notification with rollback handler
    uow.AddWithRollback(
        `CREATE nudge CONTENT { type: "new_event", ... }`,
        nudgeVars,
        func(ctx context.Context) error {
            // If transaction fails after nudge created, clean up
            return s.nudgeService.Cancel(ctx, nudgeID)
        },
    )

    return uow.Commit(ctx)
}
```

**Note:** Rollback handlers execute AFTER transaction failure, not as true ACID rollback. Use for cleanup operations.

### Pattern 4: MultiStepOperation (Sequential with Auto-Rollback)

**Location:** `internal/database/transaction.go`

Use for complex workflows where each step can fail and needs cleanup.

```go
func (s *AdventureService) CreateAdventureWorkflow(ctx context.Context, req *CreateAdventureRequest) error {
    mso := database.NewMultiStepOperation(s.db)

    var adventureID string

    mso.AddStep("create_adventure",
        func(ctx context.Context, db database.Database) error {
            adventure, err := s.repo.Create(ctx, req.Adventure)
            if err != nil {
                return err
            }
            adventureID = adventure.ID
            return nil
        },
        func(ctx context.Context, db database.Database) error {
            // Rollback: delete the adventure
            return s.repo.Delete(ctx, adventureID)
        },
    )

    mso.AddStep("create_forum",
        func(ctx context.Context, db database.Database) error {
            return s.forumRepo.CreateForAdventure(ctx, adventureID)
        },
        func(ctx context.Context, db database.Database) error {
            return s.forumRepo.DeleteByAdventureID(ctx, adventureID)
        },
    )

    mso.AddStep("notify_members",
        func(ctx context.Context, db database.Database) error {
            return s.notifyService.NotifyGuildMembers(ctx, req.GuildID, adventureID)
        },
        nil, // No rollback needed for notifications
    )

    return mso.Execute(ctx)
    // If step 2 fails, step 1's rollback executes automatically
}
```

---

## SurrealDB Specifics

### Response Format Handling

SurrealDB returns responses in a nested structure that varies by query type:

```go
// Query response format
{
    "status": "OK",
    "result": [
        {"id": "user:abc", "email": "test@test.com", ...},
        {"id": "user:def", "email": "other@test.com", ...}
    ]
}

// Sometimes wrapped further
[
    {
        "status": "OK",
        "result": [...]
    }
]
```

**Parsing pattern used in repositories:**

```go
func parseUserResult(result interface{}) (*model.User, error) {
    // Handle nil
    if result == nil {
        return nil, database.ErrNotFound
    }

    // Navigate through response wrapper
    if resp, ok := result.(map[string]interface{}); ok {
        if status, ok := resp["status"].(string); ok && status == "OK" {
            if resultData, ok := resp["result"].([]interface{}); ok {
                if len(resultData) == 0 {
                    return nil, database.ErrNotFound
                }
                result = resultData[0]
            }
        }
    }

    // Handle array wrapper
    if arr, ok := result.([]interface{}); ok {
        if len(arr) == 0 {
            return nil, database.ErrNotFound
        }
        result = arr[0]
    }

    // Now result is the actual record map
    data, ok := result.(map[string]interface{})
    if !ok {
        return nil, errors.New("unexpected result format")
    }

    // Convert SurrealDB ID format
    if id, ok := data["id"]; ok {
        data["id"] = convertSurrealID(id)
    }

    // Marshal/unmarshal for type conversion
    jsonBytes, _ := json.Marshal(data)
    var user model.User
    json.Unmarshal(jsonBytes, &user)

    return &user, nil
}
```

### RecordID Type Conversion

SurrealDB returns IDs in multiple formats depending on context:

```go
func convertSurrealID(id interface{}) string {
    // Format 1: Already a string
    if str, ok := id.(string); ok {
        return str  // "user:abc123"
    }

    // Format 2: models.RecordID struct
    if rid, ok := id.(models.RecordID); ok {
        return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
    }

    // Format 3: Pointer to RecordID
    if rid, ok := id.(*models.RecordID); ok && rid != nil {
        return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
    }

    // Format 4: Map with tb/id keys
    if m, ok := id.(map[string]interface{}); ok {
        tb := extractMapField(m, "tb", "TB", "Table")
        idPart := extractMapField(m, "id", "ID")
        return tb + ":" + idPart
    }

    // Format 5: Nested {id: {String: "value"}}
    // (handled in extractMapField)

    return fmt.Sprintf("%v", id)
}
```

### Query Parameter Binding

SurrealDB uses `$variable` syntax for parameters:

```go
// Correct - parameterized query (safe from injection)
query := `SELECT * FROM user WHERE email = $email`
vars := map[string]interface{}{"email": userEmail}
result, err := db.Query(ctx, query, vars)

// WRONG - string interpolation (SQL injection risk!)
query := fmt.Sprintf("SELECT * FROM user WHERE email = '%s'", userEmail)
```

### Record References

Reference records using `<record>` type for proper ID handling:

```go
// Reference by variable
query := `SELECT * FROM <record> $id`
vars := map[string]interface{}{"id": "user:abc123"}

// Direct reference (when ID known at query time)
query := `SELECT * FROM user:abc123`
```

---

## Repository Pattern

### Standard Repository Structure

```go
type UserRepository struct {
    db database.Database
}

func NewUserRepository(db database.Database) *UserRepository {
    return &UserRepository{db: db}
}
```

### CRUD Operations

**Create:**
```go
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    query := `
        CREATE user CONTENT {
            email: $email,
            username: $username,
            hash: $hash,
            created_on: time::now(),
            updated_on: time::now()
        }
    `
    vars := map[string]interface{}{
        "email":    user.Email,
        "username": user.Username,
        "hash":     user.Hash,
    }

    result, err := r.db.Query(ctx, query, vars)
    if err != nil {
        if isUniqueConstraintError(err) {
            return fmt.Errorf("%w: email already exists", database.ErrDuplicate)
        }
        return err
    }

    // Extract created record ID and timestamps
    created, err := extractCreatedRecord(result)
    if err != nil {
        return err
    }

    user.ID = created.ID
    user.CreatedOn = created.CreatedOn
    return nil
}
```

**Read:**
```go
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    query := `SELECT * FROM <record> $id`
    result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
    if err != nil {
        if errors.Is(err, database.ErrNotFound) {
            return nil, nil  // Not found returns nil, nil (not an error)
        }
        return nil, fmt.Errorf("getting user %s: %w", id, err)
    }
    return parseUserResult(result)
}
```

**Update:**
```go
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    query := `
        UPDATE (<record> $id) SET
            email = $email,
            username = $username,
            updated_on = time::now()
    `
    return r.db.Execute(ctx, query, map[string]interface{}{
        "id":       user.ID,
        "email":    user.Email,
        "username": user.Username,
    })
}
```

**Delete:**
```go
func (r *UserRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE (<record> $id)`
    return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}
```

---

## Error Handling

### Unique Constraint Detection

```go
// Located in repository/helpers.go
func isUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }
    errStr := err.Error()
    return strings.Contains(errStr, "unique") ||
           strings.Contains(errStr, "duplicate") ||
           strings.Contains(errStr, "already exists")
}
```

**Usage:**
```go
if isUniqueConstraintError(err) {
    return fmt.Errorf("%w: email already exists", database.ErrDuplicate)
}
```

### Standard Error Wrapping

```go
func (r *EventRepository) GetByID(ctx context.Context, id string) (*model.Event, error) {
    result, err := r.db.QueryOne(ctx, query, vars)
    if err != nil {
        if errors.Is(err, database.ErrNotFound) {
            return nil, nil  // Caller handles nil result
        }
        return nil, fmt.Errorf("getting event %s: %w", id, err)
    }
    return parseResult(result)
}
```

---

## Code Examples

### Example 1: Idempotent Point Award

```go
// TryAwardPoints attempts to award points, handling duplicates gracefully
func (r *ResonanceRepository) TryAwardPoints(ctx context.Context, entry *model.ResonanceLedgerEntry) (bool, error) {
    // Attempt to create - unique constraint will reject duplicates
    err := r.AwardPointsAtomic(ctx, entry)

    if err != nil {
        if isUniqueConstraintError(err) {
            // Already awarded - not an error, just return false
            return false, nil
        }
        return false, err
    }

    return true, nil  // Successfully awarded
}
```

### Example 2: Transactional RSVP with Capacity Check

```go
func (r *RSVPRepository) CreateWithCapacityCheck(ctx context.Context, rsvp *model.UnifiedRSVP) error {
    batch := database.NewAtomicBatch()

    // 1. Check capacity (will fail transaction if exceeded)
    batch.Add(`
        LET $event = SELECT * FROM event WHERE id = $event_id;
        LET $current = SELECT count() FROM unified_rsvp
            WHERE target_type = "event"
            AND target_id = $event_id
            AND status IN ["approved", "pending"];

        IF $current[0].count >= $event[0].max_attendees THEN {
            THROW "Event is at capacity";
        };
    `, map[string]interface{}{"event_id": rsvp.TargetID})

    // 2. Create RSVP
    batch.Add(`
        CREATE unified_rsvp CONTENT {
            target_type: $target_type,
            target_id: $target_id,
            user_id: $user_id,
            status: "pending",
            role: "participant",
            created_on: time::now(),
            updated_on: time::now()
        }
    `, map[string]interface{}{
        "target_type": rsvp.TargetType,
        "target_id":   rsvp.TargetID,
        "user_id":     rsvp.UserID,
    })

    // 3. Update denormalized count (trigger handles this, but explicit for clarity)
    batch.Add(`
        UPDATE event SET attendee_count += 1 WHERE id = type::record("event", $event_id)
    `, map[string]interface{}{"event_id": rsvp.TargetID})

    return batch.Execute(ctx, r.db)
}
```

### Example 3: Using Helper Functions

```go
// Located in repository/helpers.go

// extractCreatedRecord gets ID and timestamps from CREATE result
func extractCreatedRecord(result []interface{}) (*createdRecord, error) {
    // ... see full implementation in helpers.go
}

// extractQueryResults handles SurrealDB response unwrapping
func extractQueryResults(result interface{}) ([]interface{}, bool) {
    // ... see full implementation in helpers.go
}

// parseTime handles various time formats from SurrealDB
func parseTime(v interface{}) time.Time {
    switch t := v.(type) {
    case time.Time:
        return t
    case string:
        if parsed, err := time.Parse(time.RFC3339, t); err == nil {
            return parsed
        }
        if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
            return parsed
        }
    }
    return time.Time{}
}
```

---

## Best Practices

1. **Always use parameterized queries** - Never interpolate user input into queries

2. **Use AtomicBatch for multi-statement operations** - Ensures all-or-nothing semantics

3. **Return (nil, nil) for not found** - Let callers decide if missing record is an error

4. **Wrap errors with context** - Include operation and ID in error messages

5. **Use repository helpers** - Don't duplicate result parsing logic

6. **Test with real SurrealDB** - Mock testing misses response format issues

7. **Check for unique constraints** - Handle duplicate gracefully, not as errors

---

## Related Documentation

- [SCHEMA.md](./SCHEMA.md) - Database schema, triggers, and functions
- [PERFORMANCE.md](./PERFORMANCE.md) - Query optimization and indexing
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Overall system architecture
