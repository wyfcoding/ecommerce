#!/bin/bash
set -e

TEMPLATE_DIR="deployments/template/helm/templates"

for service_dir in deployments/*; do
  service_name=$(basename "$service_dir")
  
  # Skip template and gateway (gateway might need manual check later)
  if [[ "$service_name" == "template" || "$service_name" == "gateway" || "$service_name" == "kubernetes" || "$service_name" == "istio" ]]; then
    continue
  fi

  if [ -d "$service_dir/helm/templates" ]; then
    echo "Processing $service_name..."
    
    # 1. Copy generic templates
    cp "$TEMPLATE_DIR/_helpers.tpl" "$service_dir/helm/templates/"
    cp "$TEMPLATE_DIR/configmap.yaml" "$service_dir/helm/templates/"
    cp "$TEMPLATE_DIR/hpa.yaml" "$service_dir/helm/templates/"
    cp "$TEMPLATE_DIR/networkpolicy.yaml" "$service_dir/helm/templates/"
    cp "$TEMPLATE_DIR/pdb.yaml" "$service_dir/helm/templates/"

    # 2. Update values.yaml
    VALUES_FILE="$service_dir/helm/values.yaml"
    if ! grep -q "networkPolicy:" "$VALUES_FILE"; then
      cat >> "$VALUES_FILE" <<EOF

networkPolicy:
  enabled: false
  allowExternal: true

podDisruptionBudget:
  enabled: false
  minAvailable: 1
EOF
    fi

    # 3. Standardize helper usage in templates
    # Replace specific helpers with generic ones
    # We use perl for in-place editing to avoid BSD/GNU sed differences on Mac
    find "$service_dir/helm/templates" -name "*.yaml" -type f -exec perl -i -pe 's/include "[^"]*\.fullname"/include "fullname"/g' {} +
    find "$service_dir/helm/templates" -name "*.yaml" -type f -exec perl -i -pe 's/include "[^"]*\.labels"/include "labels"/g' {} +
    find "$service_dir/helm/templates" -name "*.yaml" -type f -exec perl -i -pe 's/include "[^"]*\.selectorLabels"/include "selectorLabels"/g' {} +
    find "$service_dir/helm/templates" -name "*.yaml" -type f -exec perl -i -pe 's/include "[^"]*\.serviceAccountName"/include "serviceAccountName"/g' {} +
    find "$service_dir/helm/templates" -name "*.yaml" -type f -exec perl -i -pe 's/include "[^"]*\.name"/include "name"/g' {} +
    
    # Also replace in _helpers.tpl itself if we accidentally overwrote it? 
    # No, we copied the generic one which uses generic names.
    # But wait, if we copied generic _helpers.tpl, it defines "fullname", "labels" etc.
    # So we just need to ensure the OTHER files use these names.
    # The perl commands above do exactly that.
  fi
done

echo "Propagation complete."
