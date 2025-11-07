Adding flightplans to *vice* is the most time consuming aspect of the facility engineering process. However, this program seeks to solve that issue.

Start of with downloading the *flightplanfiller* executable. 

Before using the program, ensure that you have an API key from OpenSky (it's free). You can make an account [here](https://opensky-network.org/login) and get an API client [here](https://opensky-network.org/my-opensky/account). Once you have the API client credentials, create a `.env` file where the executable is. That file should look something like this:
```
CLIENT_ID=<client_id>
CLIENT_SECRET=<client_secret>
```

Obviously, replace "<client_id>" and "<client_secret>" with the client ID and client secret provided by OpenSky, respectivly. 

Then, create a `resources` folder inside of the folder where the executable is. After, download the openscope-airlines.json file and move it into the resources folder. Finally, run the command with the airport you want and the amount of aircraft you want. If the amount of aircraft is omited, the default value will be 50. For example, `./flightplanfiller -airport KEWR -amount 100` (or `flightplanfiller.exe` for Windows). After, `departures.json` and `arrivals.json` will be created with all of the information needed.

Each request will take around 15 seconds, so larger requests may take some time.

Exits are calculated by the aircrafts first fix. So in cases with WHITE and DIXIE that sometimes use ELVAE, WHITE or DIXIE will not show up as the exit; rather, ELVAE will. However, you can make an `exit-exeptions.json` file in a resources folder which will be able to replace the exits. An example would look like this:
```json
[
    {
        "found_exit": "ELVAE",
        "actual_exit": ["WHITE", "DIXIE"]
    }
]
```
In this file, if the exit `ELVAE` is found, then *AirplaneFetcher* will flag it an check to see if the flightplan contains either `WHITE` or `DIXIE`. If it does, it will replace the exit with that actual exit.

You can customize scratchpad rules for this program by making a `scratchpad-rules.json`file in the resources folder. The format for scratchpads look like this:

```json
{
    "rules": [
        {
            "exit": "NEION",
            "scratchpad": "NEI"
        },
        {
            "exit": "WHITE",
            "secondary_scratchpad": "WHI"
        },
        {
            "exit": "MERIT",
            "scratchpad": "MER",
            "secondary_scratchpad": "MER"
        }
    ]
}

```

