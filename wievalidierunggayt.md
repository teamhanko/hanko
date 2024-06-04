# Validierungsregeln

1. Mindestens eine Authentifizierungsmethode muss aktiviert sein.

```yaml
 email.enabled: false
 passkey.enabled: false
 password.enabled: false
 third_party.providers:
	 google:
		 enabled: false
	 ...
 saml.enabled: false

---

email.enabled: true
email.use_for_authentication: false
passkey.enabled: false
password.enabled: false
third_party.providers:
 google:
	 enabled: false
 ...
saml.enabled: false
```

2. Es muss `email.enabled` und `email.use_for_authentication = true` sein oder `passkey.optional = false` oder `password.optional = false`.

3. Wenn `passkey.enabled = false` muss entweder `username.enabled = true` && `username.optional = false` oder `email.enabled = true` && `email.optional = false`.
