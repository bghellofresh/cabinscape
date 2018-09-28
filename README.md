# CS AirBnb Calendar Sync Service

## Problem
- potential customers can book on AirBnB despite the cabin being unavailable
- the syncing process is manual and takes too much time

## Solution
- leverage AirBnB calendar syncing to publish bookings on cabinscape main site to ical consumed by airbnb

## Proof of Concept
- [working feature branch](https://github.com/bghellofresh/cabinscape/tree/feature/api)
- [similar project](https://misell.cymru/posts/airbnb-calendar-golang/)

## Technical Requirements
- API to provide an ics file to airbnb for import
    - retrieves bookings from DB
    - returns them in ical format
- Booking Consumer to consume booking events from main site
    - consumes event
    - writes to DB filling fields necessary to create ical
- Adapt main site to publish an event to RabbitMQ when booking succeeds
    - event payload must contain necessary information for creating calendar entry

## Infrastructure Requirements
- Digital Ocean Droplet access
