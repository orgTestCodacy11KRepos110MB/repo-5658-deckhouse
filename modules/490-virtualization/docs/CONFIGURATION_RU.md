---
title: "Модуль virtualization: настройки"
---

{% include module-bundle.liquid %}

> **Важно:** Для работы модуля необходим включенный модуль [cni-cilium](../021-cni-cilium/).

Также вам потребуется указать одну или несколько желаемых подсетей из которых будут выделяться IP-адреса виртуальным машинам:

```yaml
vmCIDRs:
- 10.10.10.0/24
```

Подсеть для виртуальных машин не должна совпадать с подсетью подов и подсетью сервисов

{% include module-settings.liquid %}
