package email

func (ec *EmailClient) SendAuthEmail(emailAddr string, oneTimeCode string) error {
	fromEmailAddress := ec.SenderAddr
	m := NewEmailMessage(fromEmailAddress).AddRecipients(emailAddr).SetSubject("Klar for trappeløp?").AddStringContent("Du er nesten klar. Bruk denne koden for å bekrefte din epost: \n" + oneTimeCode)
	err := ec.SendEmail(m)
	return err
}
