
# monitoring
## Add resources definition to Loki
LOKI_PATCH="[{\"op\":\"add\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"1\",\"memory\":\"1000Mi\"},\"requests\":{\"cpu\":\"500m\",\"memory\":\"750Mi\"}}}]"

echo "Press CTRL + C to abort at any time"
echo

echo "Patching Loki will force Loki to restart, which will delete all non persisted logs"
echo "If you say no, we will only patch components "
echo -n "Are you comfortable losing your logs? [y/N] "
read -r DELETE_LOGS

if [[ "${DELETE_LOGS}" == "y" || "${DELETE_LOGS}" == "Y" ]]; then
  kubectl -n monitoring patch statefulset loki --type="json" --patch "${LOKI_PATCH}"
else
  echo "If you need help to persist your logs, see this post: https://oslokommune.slack.com/archives/CV9EGL9UG/p1629718366030700"
  echo

  echo -n "Can we still patch components which doesn't have any destructive side effects? [Y/n] "
  read -r CONTINUE
  if [[ "${CONTINUE}" != "y" && "${CONTINUE}" != "Y" ]]; then
    exit 0
  fi
fi

## Add resources definition to the AlertManager
ALERTMANAGER_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"200m\",\"memory\":\"450Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"400Mi\"}}}]"

kubectl -n monitoring patch statefulset alertmanager-kube-prometheus-stack-alertmanager --type="json" --patch "${ALERTMANAGER_PATCH}"

## Add resources definition to Prometheus
PROMETHEUS_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"300m\",\"memory\":\"500Mi\"},\"requests\":{\"cpu\":\"150m\",\"memory\":\"400Mi\"}}}]"

kubectl -n monitoring patch statefulset prometheus-kube-prometheus-stack-prometheus --type="json" --patch "${PROMETHEUS_PATCH}"

## Add resources definition to Promtail
PROMTAIL_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"500m\",\"memory\":\"256Mi\"},\"requests\":{\"cpu\":\"200m\",\"memory\":\"128Mi\"}}}]"

kubectl -n monitoring patch daemonsets promtail --type="json" --patch "${PROMTAIL_PATCH}"

## Add resources definition to Prometheus operator
PROMETHEUS_OPERATOR_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"200m\",\"memory\":\"200Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}}]"

kubectl -n monitoring patch deployment kube-prometheus-stack-operator --type="json" --patch "${PROMETHEUS_OPERATOR_PATCH}"

# kube-system
## Add resources definition to the Autoscaler
AUTOSCALER_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"300m\",\"memory\":\"750Mi\"},\"requests\":{\"cpu\":\"200m\",\"memory\":\"500Mi\"}}}]"

kubectl -n kube-system patch deployment cluster-autoscaler-aws-cluster-autoscaler --type="json" --patch "${AUTOSCALER_PATCH}"

## Add resources definition to the AWS Load Balancer controller
ALB_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"200m\",\"memory\":\"256Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}}]"

kubectl -n kube-system patch deployment aws-load-balancer-controller --type="json" --patch "${ALB_PATCH}"

## Add resources definition to the EBS CSI controller
ELB_PATCH="[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/resources\",\"value\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}}]"

kubectl -n kube-system patch deployment ebs-csi-controller --type="json" --patch "${ELB_PATCH}"
