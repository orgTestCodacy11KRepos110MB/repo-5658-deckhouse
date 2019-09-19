{{- define "instance_group_machine_class" }}
  {{- $context := index . 0 }}
  {{- $ig := index . 1 }}
  {{- $zone_name := index . 2 }}
---
{{- $tpl_context := dict }}
{{- $_ := set $tpl_context "Release" $context.Release }}
{{- $_ := set $tpl_context "Chart" $context.Chart }}
{{- $_ := set $tpl_context "Files" $context.Files }}
{{- $_ := set $tpl_context "Capabilities" $context.Capabilities }}
{{- $_ := set $tpl_context "Template" $context.Template }}
{{- $_ := set $tpl_context "Values" $context.Values }}
{{- $_ := set $tpl_context "instanceGroup" $ig }}
{{- $_ := set $tpl_context "zoneName" $zone_name }}
{{ tpl ($context.Files.Get (list "cloud-providers" $context.Values.cloudInstanceManager.internal.cloudProvider.type "machine-class.yaml" | join "/")) $tpl_context }}
{{- end }}
