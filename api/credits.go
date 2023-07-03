package api

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

func UpdatePlan(userId, db *sql.DB) {

}

func UpdatePurchasedCredits(userId, db *sql.DB) {

}

func (s *Service) UpdatePlan(c *gin.Context) {

}

func (s *Service) UpdatePurchasedCredits(c *gin.Context) {

}

func GetUsableCredits(userId int, db *sql.DB) (usableCredits int, err error) {
	rows, err := db.Query(`SELECT SUM(total) AS credits_total FROM (
					SELECT SUM(amount) as total FROM credit_purchases WHERE user_id = ? GROUP BY user_id, amount 
					UNION ALL 
					SELECT credits_per_month as total FROM plans WHERE id = IFNULL((SELECT plan_id FROM plan_purchases WHERE user_id = ? AND DATE_ADD(created_at, INTERVAL 1 MONTH) > CURRENT_TIMESTAMP), 1)
				) s;`, userId, userId)

	if err != nil {
		log.Println(err)
		return
	}

	var totalCredits int
	if err = rows.Scan(&totalCredits); err != nil {
		log.Println(err)
		return
	}

	rows, err = db.Query(`SELECT SUM(cost * (input_size / length)) as credits_used FROM history 
					JOIN action_costs ON history.cost_id = action_costs.id
					JOIN actions ON action_costs.action_id = actions.id
				WHERE user_id = ?;`, userId)

	if err != nil {
		log.Println(err)
		return
	}

	var usedCredits int
	if err = rows.Scan(&usedCredits); err != nil {
		log.Println(err)
		return
	}

	return totalCredits - usedCredits, err
}
