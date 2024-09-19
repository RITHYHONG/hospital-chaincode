package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-kivik/kivik/v3"
	"log"
	"net/http"
)

var db *kivik.DB

type Appointment struct {
	ID        string `json:"id"`
	PatientID string `json:"patient_id"`
	DoctorID  string `json:"doctor_id"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Status    string `json:"status"`
}

func initCouchDB() *kivik.DB {
	client, err := kivik.New("couch", "http://admin:password@localhost:5984/")
	if err != nil {
		log.Fatalf("Failed to connect to CouchDB: %v", err)
	}
	db := client.DB(context.Background(), "hospital")
	if err != nil {
		log.Fatalf("Failed to access database: %v", err)
	}
	return db
}

func main() {
	db = initCouchDB()

	r := gin.Default()

	r.POST("/appointments", createAppointment)
	r.GET("/appointments", getAppointments)
	r.GET("/appointments/:id", getAppointmentByID)
	r.PUT("/appointments/:id", updateAppointment)
	r.DELETE("/appointments/:id", deleteAppointment)

	r.Run(":8080")
}

func createAppointment(c *gin.Context) {
	var appointment Appointment
	if err := c.BindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Put(context.Background(), appointment.ID, appointment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create appointment: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment created successfully"})
}

func getAppointments(c *gin.Context) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{},
	}
	rows, err := db.Find(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to query documents: %v", err)})
		return
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var appointment Appointment
		if err := rows.ScanDoc(&appointment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse appointment"})
			return
		}
		appointments = append(appointments, appointment)
	}

	c.JSON(http.StatusOK, appointments)
}

func getAppointmentByID(c *gin.Context) {
	appointmentID := c.Param("id")
	row := db.Get(context.Background(), appointmentID)
	if row.Err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	var appointment Appointment
	if err := row.ScanDoc(&appointment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse appointment"})
		return
	}

	c.JSON(http.StatusOK, appointment)
}

func updateAppointment(c *gin.Context) {
	appointmentID := c.Param("id")
	var updatedAppointment Appointment
	if err := c.BindJSON(&updatedAppointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	row := db.Get(context.Background(), appointmentID)
	if row.Err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	_, err := db.Put(context.Background(), appointmentID, updatedAppointment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update appointment: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment updated successfully"})
}

func deleteAppointment(c *gin.Context) {
	appointmentID := c.Param("id")
	row := db.Get(context.Background(), appointmentID)
	if row.Err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	var appointment Appointment
	if err := row.ScanDoc(&appointment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse appointment"})
		return
	}

	rev := row.Rev
	_, err := db.Delete(context.Background(), appointmentID, rev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete appointment: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment deleted successfully"})
}
