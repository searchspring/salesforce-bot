## NPS Endpoint Docs 📋

#### Endpoint starts at `/nps` with 3 manditory fields
- `name` (string)
- `email` (string)
- `website` (string)

#### and 2 interchangable fields
*If both fields are used in the same request, the feedback will be overwritten by the rating so only use ONE at a time* 
- `rating` (int)
- `feedback` (string)

#### *There is a 6th field `test` (boolean) which if set, will return a 200 but won't post the message to slack*
<hr>

### Examples 🧰
1. This would be an example to post a message with a rating `/nps?name=clientName&email=clientEmail&website=clientWebsite&rating=clientRating`
2. This would be an example to post a message with feedback `/nps?name=clientName&email=clientEmail&website=clientWebsite&feedback=clientFeedback`
3. This would be an example to test the request BUT make no post to slack `/nps?name=clientName&email=clientEmail&website=clientWebsite&test=true`

### Help/Issues/Feature Requests 🙋
If you need help with the api, have questions, or an idea for a new feature, please either: 
- Add an issue [HERE](https://github.com/searchspring/nebo/issues/new)  
- Post a message in [Development Nebo](https://searchspring.slack.com/archives/C01N8NERZ7S)

and we will get back to you as soon as possible 😀
