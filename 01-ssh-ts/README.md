# SSH lab infra

## Print name/ip table

```sh
pulumi stack output instances | jq -r '.[] | [.name, .ip] | @tsv'
```
