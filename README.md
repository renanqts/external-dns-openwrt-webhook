# ExternalDNS Webhook Provider for OpenWRT

[ExternalDNS](https://github.com/kubernetes-sigs/external-dns) is a Kubernetes add-on for automatically managing DNS records for Kubernetes ingresses and services by using different DNS providers. This webhook provider allows you to automate DNS records from your Kubernetes clusters into your OpenWRT router. If you like home automation like me, it should help you.

For examples of creating DNS records either via CRDs or via Ingress/Service annotations, check out the [example directory](./example).

## Limitations
- `DNSEndpoints` with multiple `targets` are not supported.
- Supported DNS record types: `A`, `CNAME`.
- Only `psert-only` policy is supported.

## Configuration Options
You can find all the environment variables allowed as well as the default in the [values file](example/values.yaml#L19).   
The installation can be achieved via [helm chart](skaffold.yaml#L15-L26).

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/renanqts4)
