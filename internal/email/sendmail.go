package email

func (ec *EmailClient) SendAuthEmail(emailAddr string, oneTimeCode string) error {
	fromEmailAddress := ec.SenderAddr
	m := NewEmailMessage(fromEmailAddress).AddRecipients(emailAddr).SetSubject("Klar for trappeløp?").AddStringContent("Du er nesten klar. Bruk denne koden for å bekrefte din epost: \n" + oneTimeCode)
	err := ec.SendEmail(m)
	return err
}

func (ec *EmailClient) SendPasswordCode(code string, toEmail string) error {
	fromEmailAddress := ec.SenderAddr
	m := NewEmailMessage(fromEmailAddress).AddRecipients(toEmail).SetSubject("Tilbakestill ditt passord").AddStringContent("Bruk denne koden for å tilbakestille ditt passord: \n" + code)
	err := ec.SendEmail(m)
	return err
}
