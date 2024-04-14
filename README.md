Adding flightplans to *vice* is the most time consuming aspect of the facility engineering process. However, this program seeks to solve that issue.


Start of with downloading the *AirplaneFetcher* executable. Then, download the openscope-airlines.json file and move it into the same folder as the executable. Finally, run the command with the airport you want and the amount of aircraft you want. If the amount of aircraft is omited, the default value will be 50. For example, `./AirplaneFetcher -airport KEWR -amount 100`. After, `departures.json` and `arrivals.json` will be created with all of the information needed.

Please be aware that each request will take around 15 seconds, so large amounts of aircraft will take some time.

You can customize scratchpad rules for this program by making a `scratchpadRules.json` file. The format for scratchpads look like this:

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