# IMP: Microblogging Platform

IMP is a Twitter-like service that is designed to be decentralized. In the same way that no single business can own all email or all websites, no single business can own all IMP accounts. Anyone can run an IMP host, and users on different hosts can connect with each other, because IMP is an open standard.

## Notes

IMP is a system for posting *notes*. Notes are basically the same as tweets on Twitter, but for obvious reasons we're not calling them tweets.

Notes are limited to 140 characters in length, however, you do get a few freebies:

* The longest URL counts only for 2 characters.
* @-mentions of other users only count for 2 characters, *if* the other user allows you to @-mention them.
* Replies are natively supported without requiring you to @-mention the original poster, *if* you are allowed to reply.
* Re-posts are free. You have the entire 140 characters to comment on a note.
* Hashtags may be supported somehow, but they are counted for their full length.
* A note can include a free "hat tip" or "via" mention, if you are allowed to mention the other user.

Shortening of URLs and @-mentions is not a way to sneak in extra text, as only valid @-mentions get shortened, and client applications might not display the entire URL.

Notes can be edited or deleted. Edited notes are flagged as such.

### Status and Essence

We settle the old Jack vs. Ev / status vs. messaging debate by providing that your user profile may contain a *status*, which has all the features of a note, but you only have one, and old stati are not archived.

Your profile can also have a biography, called an *essence* (for linguistic reasons that really aren't relevant), which has the properties of a note. You only have one, and it's not archived. It's meant to be a more permanent description, while status is meant for what you're doing *right now*.

### DESIGN DEBATE!

None of these policies are written in stone. I'd like to hear your opinions about   what features should be supported by notes.

There are also some tricky implementation details about where notes live in a decentralized service. For example, is a re-post just a link to the original, or is the content of the note copied to the reposter's host?

## Handles

Usernames look like email addresses, however, to avoid confusion they use "!" instead of "@". The *handle* comes before the "!" and consists of 16 characters, which may be any combination "_" (underscore), "0" through "9", "a" through "z", or "A" through "Z". Handles will always be displayed with the case chosen by the user, but they will be treated as case-insensitive, therefore "example!example.com" and "eXampLe!example.com" both refer to the same user.

The *host* portion of the address follows the "!" and is a fully-qualified domain name without the trailing dot. When there is no ambiguity, client applications can hide the host portion of an address.

The word "handle" can refer to the first part of the address or the whole thing.

### DESIGN DEBATE!

Is "!" the best choice for handles? I don't want to use "@" or "#", for obvious reasons. Should the host come before the handle?

## Groups

Groups are like "circles" in Google+. Details TBA.

## Blocking and Muting

Two overarching design principles:

* A muted user can see you, but you have no idea that they exist, except when you look at your mute list.
* A blocked user is also like a muted user, but additionally has no idea that you exist.

Corollaries to the above:

* @-mentions of you by blocked users won't be parsed or shortened, because from their point of view, you are not an existing user.
* Muted users can @-mention you, but you will not be notified and you will not see them.

Additional proposed policies, some of these could be user-configurable:

* Non-followers are muted.
* New followers are muted for the first 2 weeks.
* You can block or mute everyone at a host, e.g. *!example.com.

### DESIGN DEBATE!

I've never had to deal with harassment, so I'm not very familiar with what forms it can take, how it is perpetrated, and how it is effectively dealt with. I know a lot of people have been working on how to deal with harassment on Twitter. I'd like to hear their ideas.

One disadvantage of creating a decentralized service is that we can't enforce content policies on every host of that service. (Blogger can set rules for the blogs they host, but they have no control over blogs hosted elsewhere.) I hope we can structure the service so that users can control what they see, while still feeling like everyone is on the same network.

## Additional Features

These are things I'd like to include, but haven't figured out the best way to do it yet:

* Direct messages
* Favorites
* Bookmarks (like a private favorite)
* Data export and import in a standard format
* Public key encrypted direct messages, not even host can read

# Implementation

What follows is more technical stuff.

## Service Discovery

Usernames take the form of *handle!host*. However, it may not be convenient for the IMP service to run at the root of the host. Therefore a service discovery mechanism is proposed.

Given the host name imp.example.com, a client searches for the service at these URLs in order:

1. https://imp.example.com/
2. https://imp.example.com:5039/
3. http://imp.example.com/ (Note that this request is for discovery only. All IMP API requests must use HTTPS.)

IMP service providers should respond as early in the search as possible using one of the following:

1. Include this HTTP header:

    `IMP-API-Location: 0.9;imp.example.com/path/to/api`
    
2. Do a 301 redirect to the API location.
3. Return an HTML page with the following meta tag in its head element:

    `<meta http-equiv="IMP-API-Location" content="0.9;imp.example.com/path/to/api" />`

Every response from an IMP service must include the `IMP-API-Location` header. 

## Guest Authentication

An IMP server has *users* and *guests*. Users are people whose accounts are hosted on the IMP server. They authenticate directly with their host to manage their accounts and post notes.

Guests are people whose accounts are hosted on other systems but who want to follow users on this system. If guests were identified solely by handle and hostname, then they could easily circumvent block lists or join private groups by spoofing handles. Therefore, guests must participate in *guest authentication*, which works as follows:

1. alice!host.a wants to follow bob!host.b
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

GET /user/{handle}/host/{host}

Get user's guest auth token for the host.

POST /guest

Called on foreign host by a user's host to create a guest token.

POST /user/{handle}/host

Called on user's host by a foreign host to return a guest token.

### Users

POST /user

Create a new user.

### Notes

GET /note

Get the authenticated user's notes. *NO!? That introduces state. The user should be a query parameter.*

POST /note

Create a new note from the authenticated user. *Author should be a field in the note object, otherwise we're violating statelessness.*

GET /note/{id}

Retrieve the specified note.

PUT /note/{id}

Edit the text of an existing note.

DELETE /note/{id}

Delete the specified note.

### Groups

GET /group

List groups. Users can only see the groups they own.

POST /group

Create a group with a name for a user.

GET /group/{id}

List the members of a group.

PUT /group/{id}

Edit group properties.

DELETE /group/{id}

Delete the group.

PUT /group/{id}/{address}

Add the address to the group.

DELETE /group/{id}/{address}

Remove the address from the group.

### Mutes and Blocks

GET /user/{handle}/mute

List all the user's mutes. Only the authenticated user can see these.

PUT /user/{handle}/mute/{address}

Mute the user at *address*.

DELETE /user/{handle}/mute/{address}

Unmute the user at *address*.

GET /user/{handle}/block

List all the user's blocks. Only the authenticated user can see these.

PUT /user/{handle}/block/{address}

Block the user at *address*.

DELETE /user/{handle}/block/{address}

Unblock the user at *address*.

## To-Do List

* Ports to other languages and platforms.
* A compliance suite for testing that IMP instances conform to the specification.
* Twitter bridge: Tweets your IMP notes, posts your tweets to IMP.
* Client SDKs for iOS and Android.
