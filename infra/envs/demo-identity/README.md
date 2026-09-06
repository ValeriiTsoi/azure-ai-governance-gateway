# Demo Identity

This Terraform root manages the Microsoft Entra ID identities used by the Azure AI Governance Gateway demo.

It owns:

- Governance API application registration;
- Governance API service principal;
- public demo client application registration;
- demo client service principal;
- delegated `access_as_user` OAuth2 scope.

The state is isolated from the main Azure infrastructure state:

```text
demo-identity.tfstate
```

No client secret is created or stored.

The demo client is configured as a public client and can use an interactive/device-code OAuth flow to obtain a delegated access token for the Governance API.

The Governance API audience is:

```text
api://aigov-governance-api-demo
```

Stage 7 additionally supports a separate API-key-style **gateway credential** at APIM for IDE/API clients that cannot perform interactive Microsoft Entra authentication. That credential is stored in Azure Key Vault and is not an Azure OpenAI key.
