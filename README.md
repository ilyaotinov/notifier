# Simple notification app.

## Description
App is a deamon, that runs HTTP server for recive notifications for futher sending to telegram.

## HTTP server endpoints
### [POST] /notification
Create new notification with specified title. It will be sended to telegram after specified scheduled_at datetime

```json 
{
	title: "string",
	scheduled_at: "datetime"
}
```

### [GET] /notification?with_sended=true
return list of user notifications. If with_sended parameter equal to true, then include in list
already sended notificatons. Otherwise list will contains only notifications which is about to be sended.
```json
{
	notifications: [
		{
			title: "string"
			scheduled_at: "datetime",
			sended_at: "datetime",
			created_at: "datetime"
		},
		...
	]
}
```
### [POST] /user/reigser
```json
	{
		login: "string",
		password: "string"
	}
```

## Deamon sender
This part of application is retrive from internal storage notifications where sended_at field is NULL and sends it to
telegram, if scheduled_at <= now(). After that it updated sended_at filed to now().

## Registraion
For connect between cli app with reminders, telegram and specific user, there is a flow of registration:
1. User from cli sends [POST] /user/register request.
2. User from telegram bot, using /start command, specifies login from first step, and connect self telegram chat to notifier.
3. After that, user can recieve notifications.

