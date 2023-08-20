package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func updatePlan(userId, db *sql.DB) {

}

func updatePurchasedCredits(userId, db *sql.DB) {

}

func (s *Service) UpdatePlan(c *gin.Context) {

}

func (s *Service) UpdatePurchasedCredits(c *gin.Context) {

}

func getUsableCredits(userId int, db *sql.DB) (usableCredits int, err ClientError) {
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
