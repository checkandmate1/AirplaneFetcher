Adding flightplans to *vice* is the most time consuming aspect of the facility engineering process. However, this program seeks to solve that issue.

Before you use this tool, you are going to need a free API key from https://aviationstack.com/. With this API, you will be allowed 100 free API requests per month. 

Start of with downloading the *AirplaneFetcher* executable. Then, download the openscope-airlines.json file and move it into the same folder as the executable. Finally, make a `.env` file in the same folder. In the `.env` file, write `API_KEY = "[insert_your_API_key_here]"`. Finally, run the command with the airport you want. For example `./AirplaneFetcher -airport KEWR`. After, an `output.json` file will appear will all of the relavent information.

You can customize scratchpad rules for this program by making a `scratchpadRules.json` file. The format for scratchpads look like this:
```json
{
    "rules": [
        {
            "exit": "GAYEL",
            "scratchpad": "GAY"
        },
        {
            "exit": "WAVEY",
            "scratchpad": "WAV"
        },
        {
            "exit": "NEION",
            "scratchpad": "NEI"
        },
        {
            "exit": "WHITE",
            "scratchpad": "WHI"
        },
        {
            "exit": "MERIT",
            "scratchpad": "MER",
            "secondary_scratchpad": "MER"
        }
    ]
}

```

NOTE: This only works for departure flightplans for now. Arrivals will be added soon™