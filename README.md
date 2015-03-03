# IMP: Microblogging Platform

IMP is a Twitter-like service that is designed to be decentralized. In the same way that no single business can own all email or all websites, no single business can own all IMP accounts. Anyone can run an IMP host, and users on different hosts can connect with each other, because IMP is an open standard.

<!-- TODO: Later:
## Service Discovery

Usernames take the form of *handle*@*host*. However, it may not be convenient for the IMP service to run at the root of the host. Therefore a service discovery mechanism is proposed.

For any IMP instance to be considered a host, the response to a GET request on the root path of the host must be an HTML page that contains the following LINK element in its HEAD element:

    <link rel="imp-api-prefix" href="/prefix"/>

The value of the `href` attribute will be prepended onto all IMP API calls. This attribute may be an empty string.
-->

## Guest Authentication

An IMP server has *users* and *guests*. Users are people whose accounts are hosted on the IMP server. They authenticate directly with their host to manage their accounts and post notes.

Guests are people whose accounts are hosted on other systems but who want to follow users on this system. If guests were identified solely by handle and hostname, then they could easily circumvent block lists or join private groups by spoofing handles. Therefore, guests must participate in *guest authentication*, which works as follows:

1. alice@host.a wants to follow bob@host.b
2. Alice GETs the token from /user/alice/host/host.b on host.a
3. The token doesn't exist, so host.a returns 202 Accepted
4. host.a creates a random nonce and POSTs Alice's address and the nonce to /guest on host.b
5. host.b creates an auth token and POSTs it and the nonce at /user/alice/host on host.a
6. host.a verifies the nonce and stores the token
7. Alice retries her GET /user/alice/host/host.b on host.a and this time receives the token
8. Alice can now query host.b for Bob's notes. She must supply her token in every interation with host.b

This process is similar to how you supply your email address when creating an account on a website. The website doesn't simply trust that you own the email address; it sends a verification code *to the address* so that you can click on the link in the email to prove you own the address. If you tried to use someone else's address, you would never see the verifaction email.

## HTTPS

All API calls **must** use HTTPS. Any calls to an IMP service over unencrypted HTTP will be redirected to the root of the domain. They will not simply be redirected to the same URL with an https scheme, as this would encourage continued use of unencrypted HTTP for the initial request.

## API

### Authentication

POST /token

Post credentials and get an auth token.

DELETE /token/{token}

Delete an auth token.

### Users

POST /user

Create a new user.

### Notes

GET /note

Get the authenticated user's notes. 

POST /note

Create a new note from the authenticated user.

GET /note/{id}

Retrieve the specified note.

PUT /note/{id}

Edit the text of an existing note.

DELETE /note/{id}

Delete the specified note.

## To-Do List

* Ports to other languages and platforms.
* A compliance suite for testing that IMP instances conform to the specification.
