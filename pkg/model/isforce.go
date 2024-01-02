package model

import "gorm.io/gorm"

func isForce(tx *gorm.DB) bool {
	force, ok := tx.InstanceGet("force")
	return ok && force.(bool)
}
