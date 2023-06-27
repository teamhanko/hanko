# Adding OIDC Clients
To successfully test this you need to add OIDC Clients as Secrets:

Create a github.env and a google.env of the form:
```
client_id=your-id
client_secret=your-secret
```

Run
> skaffold run -p thirdparty-x-domain

to build and deploy to local cluster.

The quickstart app should then be running on **https://app.domain-app.grocery**
