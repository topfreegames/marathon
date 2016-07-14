package models_test

//
// import (
// 	"git.topfreegames.com/topfreegames/marathon/models"
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/satori/go.uuid"
// )
//
// var _ = Describe("Models", func() {
// 	var (
// 		db models.DB
// 	)
// 	BeforeEach(func() {
// 		_db, dbErr := models.GetTestDB()
// 		Expect(dbErr).To(BeNil())
// 		Expect(_db).NotTo(BeNil())
// 		db = _db
// 	})
//
// 	Describe("UserToken", func() {
// 		Describe("Create UserToken", func() {
// 			It("Should create a template through a factory", func() {
// 				template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 				Expect(templateErr).To(BeNil())
// 				insertTemplateErr := db.Insert(template)
// 				Expect(insertTemplateErr).To(BeNil())
//
// 				dbTemplate, dbTemplateErr := models.GetTemplateByID(db, template.ID)
// 				Expect(dbTemplateErr).To(BeNil())
// 				Expect(dbTemplate.Name).To(Equal(template.Name))
// 				Expect(dbTemplate.Locale).To(Equal(template.Locale))
// 				Expect(dbTemplate.Defaults).To(Equal(template.Defaults))
// 				Expect(dbTemplate.Body).To(Equal(template.Body))
// 			})
//
// 			It("Should create a template", func() {
// 				name := uuid.NewV4().String()
// 				service := uuid.NewV4().String()[:4]
// 				locale := uuid.NewV4().String()[:2]
// 				defaults := map[string]interface{}{"username": "banduk"}
// 				body := map[string]interface{}{"alert": "{{username}} sent you a message."}
// 				createdTemplate, createdTemplateErr := models.CreateTemplate(db, name, service, locale, defaults, body)
// 				Expect(createdTemplateErr).To(BeNil())
//
// 				dbTemplate, dbTemplateErr := models.GetTemplateByID(db, createdTemplate.ID)
// 				Expect(dbTemplateErr).To(BeNil())
// 				Expect(dbTemplate.Name).To(Equal(createdTemplate.Name))
// 				Expect(dbTemplate.Locale).To(Equal(createdTemplate.Locale))
// 				Expect(dbTemplate.Defaults).To(Equal(createdTemplate.Defaults))
// 				Expect(dbTemplate.Body).To(Equal(createdTemplate.Body))
// 			})
//
// 			It("Should not create a template with repeated name,locale", func() {
// 				name := uuid.NewV4().String()
// 				service := uuid.NewV4().String()[:4]
// 				locale := uuid.NewV4().String()[:2]
// 				defaults1 := map[string]interface{}{"username": "banduk"}
// 				body1 := map[string]interface{}{"alert1": "{{username1}} sent you a message1."}
// 				defaults2 := map[string]interface{}{"username": "banduk"}
// 				body2 := map[string]interface{}{"alert2": "{{username2}} sent you a message2."}
//
// 				_, createdTemplateErr1 := models.CreateTemplate(db, name, service, locale, defaults1, body1)
// 				Expect(createdTemplateErr1).To(BeNil())
//
// 				_, createdTemplateErr2 := models.CreateTemplate(db, name, service, locale, defaults2, body2)
// 				Expect(createdTemplateErr2).NotTo(BeNil())
// 			})
// 		})
// 	})
// 	// Describe("Update Template", func() {
// 	// 	It("Should update a template for an existent id", func() {
// 	// 		template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr).To(BeNil())
// 	// 		insertTemplateErr := db.Insert(template)
// 	// 		Expect(insertTemplateErr).To(BeNil())
//   //
// 	// 		name := uuid.NewV4().String()
// 	// 		service := uuid.NewV4().String()[:4]
// 	// 		locale := uuid.NewV4().String()[:2]
// 	// 		defaults := map[string]interface{}{"username": "banduk"}
// 	// 		body := map[string]interface{}{"alert": "{{username}} sent you a message."}
// 	// 		updatedTemplate, updatedTemplateErr := models.UpdateTemplate(db, template.ID, name, service, locale, defaults, body)
// 	// 		Expect(updatedTemplateErr).To(BeNil())
//   //
// 	// 		dbTemplate, dbTemplateErr := models.GetTemplateByID(db, template.ID)
// 	// 		Expect(dbTemplateErr).To(BeNil())
// 	// 		Expect(updatedTemplate.Name).To(Equal(dbTemplate.Name))
// 	// 		Expect(updatedTemplate.Name).To(Equal(name))
// 	// 		Expect(updatedTemplate.Locale).To(Equal(dbTemplate.Locale))
// 	// 		Expect(updatedTemplate.Locale).To(Equal(locale))
// 	// 		Expect(updatedTemplate.Defaults).To(Equal(dbTemplate.Defaults))
// 	// 		Expect(updatedTemplate.Defaults).To(Equal(defaults))
// 	// 		Expect(updatedTemplate.Body).To(Equal(dbTemplate.Body))
// 	// 		Expect(updatedTemplate.Body).To(Equal(body))
// 	// 	})
//   //
// 	// 	It("Should not update a template with repeated name,locale,service", func() {
// 	// 		template1, templateErr1 := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr1).To(BeNil())
// 	// 		insertTemplateErr1 := db.Insert(template1)
// 	// 		Expect(insertTemplateErr1).To(BeNil())
//   //
// 	// 		template2, templateErr2 := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr2).To(BeNil())
// 	// 		insertTemplateErr2 := db.Insert(template2)
// 	// 		Expect(insertTemplateErr2).To(BeNil())
//   //
// 	// 		defaults := map[string]interface{}{"username": "banduk"}
// 	// 		body := map[string]interface{}{"alert": "{{username}} sent you a message."}
// 	// 		_, updatedTemplateErr := models.UpdateTemplate(db, template2.ID, template1.Name, template1.Service, template1.Locale, defaults, body)
// 	// 		Expect(updatedTemplateErr).NotTo(BeNil())
// 	// 		dbTemplate, dbTemplateErr := models.GetTemplateByID(db, template2.ID)
// 	// 		Expect(dbTemplateErr).To(BeNil())
// 	// 		Expect(template2.Name).To(Equal(dbTemplate.Name))
// 	// 		Expect(template2.Locale).To(Equal(dbTemplate.Locale))
// 	// 		Expect(template2.Defaults).To(Equal(dbTemplate.Defaults))
// 	// 		Expect(template2.Body).To(Equal(dbTemplate.Body))
// 	// 	})
//   //
// 	// 	It("Should not update a template for an unexistent id", func() {
// 	// 		template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr).To(BeNil())
// 	// 		insertTemplateErr := db.Insert(template)
// 	// 		Expect(insertTemplateErr).To(BeNil())
//   //
// 	// 		defaults := map[string]interface{}{"username": "banduk"}
// 	// 		body := map[string]interface{}{"alert": "{{username}} sent you a message."}
// 	// 		invalidID := uuid.NewV4()
// 	// 		_, updatedTemplateErr := models.UpdateTemplate(db, invalidID, template.Name, template.Service, template.Locale, defaults, body)
// 	// 		Expect(updatedTemplateErr).NotTo(BeNil())
// 	// 	})
// 	// })
//   //
// 	// Describe("Get Template", func() {
// 	// 	It("Should retrieve a template for an existent id", func() {
// 	// 		template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr).To(BeNil())
// 	// 		insertTemplateErr := db.Insert(template)
// 	// 		Expect(insertTemplateErr).To(BeNil())
//   //
// 	// 		dbTemplate, dbTemplateErr := models.GetTemplateByID(db, template.ID)
// 	// 		Expect(dbTemplateErr).To(BeNil())
// 	// 		Expect(dbTemplate.Name).To(Equal(template.Name))
// 	// 		Expect(dbTemplate.Locale).To(Equal(template.Locale))
// 	// 		Expect(dbTemplate.Defaults).To(Equal(template.Defaults))
// 	// 		Expect(dbTemplate.Body).To(Equal(template.Body))
// 	// 	})
//   //
// 	// 	It("Should not retrieve a template for an unexistent id", func() {
// 	// 		invalidID := uuid.NewV4()
// 	// 		_, dbTemplateErr := models.GetTemplateByID(db, invalidID)
// 	// 		Expect(dbTemplateErr).NotTo(BeNil())
// 	// 	})
//   //
// 	// 	It("Should retrieve all templates for an existent name", func() {
// 	// 		templates := []*models.Template{}
//   //
// 	// 		template1, templateErr1 := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr1).To(BeNil())
// 	// 		templates = append(templates, template1)
// 	// 		insertTemplateErr1 := db.Insert(template1)
// 	// 		Expect(insertTemplateErr1).To(BeNil())
//   //
// 	// 		name := template1.Name
// 	// 		service := uuid.NewV4().String()[:4]
// 	// 		locale := uuid.NewV4().String()[:2]
// 	// 		defaults := map[string]interface{}{"username": "banduk"}
// 	// 		body := map[string]interface{}{"alert": "{{username}} sent you a message."}
// 	// 		template2, templateErr2 := models.CreateTemplate(db, name, service, locale, defaults, body)
// 	// 		Expect(templateErr2).To(BeNil())
// 	// 		templates = append(templates, template2)
//   //
// 	// 		dbTemplates, dbTemplatesErr := models.GetTemplatesByName(db, template1.Name)
// 	// 		Expect(dbTemplatesErr).To(BeNil())
// 	// 		Expect(len(dbTemplates)).To(Equal(2))
// 	// 		for index, dbTemplate := range dbTemplates {
// 	// 			Expect(dbTemplate.Name).To(Equal(templates[index].Name))
// 	// 		}
// 	// 	})
//   //
// 	// 	It("Should not retrieve templates for an unexistent name", func() {
// 	// 		invalidName := uuid.NewV4().String()
// 	// 		_, dbTemplateErr := models.GetTemplatesByName(db, invalidName)
// 	// 		Expect(dbTemplateErr).NotTo(BeNil())
// 	// 	})
//   //
// 	// 	It("Should retrieve a template for existent name,service,locale", func() {
// 	// 		template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr).To(BeNil())
// 	// 		insertTemplateErr := db.Insert(template)
// 	// 		Expect(insertTemplateErr).To(BeNil())
//   //
// 	// 		dbTemplate, dbTemplateErr := models.GetTemplateByNameServiceAndLocale(db, template.Name, template.Service, template.Locale)
// 	// 		Expect(dbTemplateErr).To(BeNil())
// 	// 		Expect(dbTemplate.Name).To(Equal(template.Name))
// 	// 		Expect(dbTemplate.Locale).To(Equal(template.Locale))
// 	// 		Expect(dbTemplate.Defaults).To(Equal(template.Defaults))
// 	// 		Expect(dbTemplate.Body).To(Equal(template.Body))
// 	// 	})
//   //
// 	// 	It("Should not retrieve a template for invalid name,service,locale", func() {
// 	// 		template, templateErr := CreateTemplateFactory(db, map[string]interface{}{})
// 	// 		Expect(templateErr).To(BeNil())
// 	// 		insertTemplateErr := db.Insert(template)
// 	// 		Expect(insertTemplateErr).To(BeNil())
//   //
// 	// 		invalidLocale := uuid.NewV4().String()[:2]
// 	// 		_, dbTemplateErr := models.GetTemplateByNameServiceAndLocale(db, template.Name, template.Service, invalidLocale)
// 	// 		Expect(dbTemplateErr).NotTo(BeNil())
// 	// 	})
// 	// })
// })
