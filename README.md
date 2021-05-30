# cowin-alerts
Sends a Pager duty alert when vaccines on the Cowin website (for people aged 18-45) are available for a given list of center codes

* Before running, make sure environment variables are set as mentioned in the env.sample file. Make sure you add center codes for which you want alerts.
* Make sure you add your Pager Duty key to the environment variables (as mentioned in env.sample).
* Install the Pager Duty app and set it up with the same account you used to generate the Pager Duty key.
* As long as the app runs, you'll get an alert when a center with the center code you've mentioned in the .env file has vaccination for the 18-45 age group available (first dose).
