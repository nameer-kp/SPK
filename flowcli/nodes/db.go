package nodes

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nameer-kp/flowcli/pkg/node"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DBNode executes database operations
type DBNode struct{}

func NewDBNode() *DBNode {
	return &DBNode{}
}

func (n *DBNode) Name() string {
	return "db"
}

func (n *DBNode) Description() string {
	return "Execute database queries"
}

func (n *DBNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"connection": "string", // DSN or profile reference
		"driver":     "string", // postgres, mysql, sqlite
		"operation":  "string", // query, exec
		"sql":        "string",
		"params":     "array",
	}
}

func (n *DBNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	driver, _ := config["driver"].(string)
	if driver == "" {
		driver = "postgres"
	}

	connection, ok := config["connection"].(string)
	if !ok || connection == "" {
		// Try to get from profile
		if conn, ok := ctx.GetProfileVar("db_connection"); ok {
			connection = conn
		} else {
			return node.Result{}, fmt.Errorf("connection is required")
		}
	}

	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "query"
	}

	sqlStr, ok := config["sql"].(string)
	if !ok || sqlStr == "" {
		return node.Result{}, fmt.Errorf("sql is required")
	}

	// Extract params
	var params []interface{}
	if paramsRaw, ok := config["params"].([]interface{}); ok {
		params = paramsRaw
	}

	ctx.Logger().Info("executing database operation", "driver", driver, "operation", operation)

	// Connect to database
	db, err := sqlx.Connect(driver, connection)
	if err != nil {
		return node.Result{
			Success: false,
			Status:  "Connection failed",
			Logs:    []node.LogEntry{{Level: "error", Message: err.Error()}},
		}, err
	}
	defer db.Close()

	switch operation {
	case "query":
		return n.query(ctx, db, sqlStr, params)
	case "exec":
		return n.execSQL(ctx, db, sqlStr, params)
	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *DBNode) query(ctx node.Context, db *sqlx.DB, sqlStr string, params []interface{}) (node.Result, error) {
	rows, err := db.Queryx(sqlStr, params...)
	if err != nil {
		return node.Result{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return node.Result{}, fmt.Errorf("scan failed: %w", err)
		}
		// Convert []byte to string for JSON serialization
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return node.Result{}, fmt.Errorf("row iteration error: %w", err)
	}

	return node.Result{
		Success: true,
		Data: map[string]interface{}{
			"rows":  results,
			"count": len(results),
		},
		Status: fmt.Sprintf("%d rows", len(results)),
		Logs:   []node.LogEntry{{Level: "info", Message: fmt.Sprintf("Query returned %d rows", len(results))}},
	}, nil
}

func (n *DBNode) execSQL(ctx node.Context, db *sqlx.DB, sqlStr string, params []interface{}) (node.Result, error) {
	result, err := db.Exec(sqlStr, params...)
	if err != nil {
		return node.Result{}, fmt.Errorf("exec failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	return node.Result{
		Success: true,
		Data: map[string]interface{}{
			"rows_affected":  rowsAffected,
			"last_insert_id": lastInsertID,
		},
		Status: fmt.Sprintf("%d affected", rowsAffected),
		Logs:   []node.LogEntry{{Level: "info", Message: fmt.Sprintf("Exec affected %d rows", rowsAffected)}},
	}, nil
}

// Ensure unused import doesn't cause issues
var _ = sql.ErrNoRows
