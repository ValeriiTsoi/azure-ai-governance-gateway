# Demo Identity

This Terraform root manages the Microsoft Entra ID identities used by the Azure AI Governance Gateway demo.

It owns:

- Governance API application registration
- Governance API service principal
- Public demo client application registration
- Demo client service principal
- Delegated `access_as_user` OAuth2 scope

The state is isolated from Azure infrastructure resources:

~~~text
demo-identity.tfstate
~~~

No client secret is created or stored.

The demo client is configured as a public client and will later use an interactive OAuth flow to obtain an access token for the Governance API.

The Governance API audience is:

~~~text
api://aigov-governance-api-demo
~~~
