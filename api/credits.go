package api

import (
	"database/sql"
	"net/http"
)

func updatePlan(userId int, planId int, stripeTransactionId string, db *sql.DB) ClientError {
	row := db.QueryRow("SELECT email FROM users WHERE id = ?;", userId)

	var email string
	if err := row.Scan(&email); err != nil {
		return NewHttpError(err, http.StatusBadRequest, "This user doesn't exist")
	}

	row = db.QueryRow("SELECT plan_id FROM plan_purchases s WHERE user_id = ? AND DATE_ADD(s.created_at, INTERVAL 1 MONTH) > CURRENT_TIMESTAMP;", userId)

	var currentPlanId int
	if err := row.Scan(&currentPlanId); err == nil {
		return NewHttpError(err, http.StatusBadRequest, "This user is already subscribed to a paid plan")
	}

	insertStmt, err := db.Prepare("INSERT into plan_purchases (user_id, plan_id, stripe_transaction_id) VALUES (?, ?, ?);")
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}
	defer insertStmt.Close()

	insertResult, err := insertStmt.Exec(userId, planId, stripeTransactionId)
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	_, err = insertResult.LastInsertId()
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	return err
}

func updatePurchasedCredits(userId int, subtotal int64, stripeTransactionId string, db *sql.DB) ClientError {
	row := db.QueryRow("SELECT email FROM users WHERE id = ?;", userId)

	var email string
	if err := row.Scan(&email); err != nil {
		return NewHttpError(err, http.StatusBadRequest, "This user doesn't exist")
	}

	insertStmt, err := db.Prepare("INSERT into credit_purchases (user_id, amount, stripe_transaction_id) VALUES (?, ?, ?);")
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}
	defer insertStmt.Close()

	insertResult, err := insertStmt.Exec(userId, float64(subtotal)/100, stripeTransactionId)
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	_, err = insertResult.LastInsertId()
	if err != nil {
		return NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	return err
}

func getUsableCredits(userId int, db *sql.DB) (usableCredits float64, err ClientError) {
	row := db.QueryRow(`SELECT SUM(credits) AS usable_credits FROM (
				SELECT SUM(total) AS credits FROM (
					SELECT SUM(amount) as total FROM credit_purchases WHERE user_id = ? GROUP BY user_id, amount 
					UNION ALL 
					SELECT credits_per_month as total FROM plans WHERE id = IFNULL((SELECT plan_id FROM plan_purchases WHERE user_id = ? AND DATE_ADD(created_at, INTERVAL 1 MONTH) > CURRENT_TIMESTAMP), 1)
				) s UNION ALL 
				SELECT -IFNULL(SUM(cost * (input_size / length)), 0) as credits FROM history 
					JOIN action_costs ON history.cost_id = action_costs.id
					JOIN actions ON action_costs.action_id = actions.id
				WHERE user_id = ? AND status != 2
			) s;`, userId, userId, userId)

	if err = row.Scan(&usableCredits); err != nil {
		return 0, NewHttpError(err, http.StatusInternalServerError, "There was a problem with the server")
	}

	return usableCredits, err
}
