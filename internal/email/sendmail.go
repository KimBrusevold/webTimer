package email

func (ec *EmailClient) SendAuthEmail(emailAddr string) error {
	fromEmailAddress := ec.SenderAddr
	m := NewEmailMessage(fromEmailAddress).AddRecipients(emailAddr).SetSubject("Klar for trappeløp?").AddStringContent("Her er din unike kode: 123123123")
	err := ec.SendEmail(m)
	return err
}
