export interface SafeguardResult {
  snippet: string;
  guidance: string;
}

export function generateHCLSafeguard(item: {
  address?: string;
  resource_address?: string;
  action?: string;
  is_stateful?: boolean;
  replacement_keys?: string[];
  category?: string;
}): SafeguardResult {
  const lines: string[] = ['  lifecycle {'];

  if (item.is_stateful) {
    lines.push('    prevent_destroy       = true');
  }

  if (item.action === 'replace') {
    lines.push('    create_before_destroy = true');
  }

  if (item.replacement_keys && item.replacement_keys.length > 0) {
    lines.push(`    # Attribute changes forcing recreation: ${item.replacement_keys.join(', ')}`);
    lines.push(`    ignore_changes = [${item.replacement_keys.map(k => `"${k}"`).join(', ')}]`);
  }

  lines.push('  }');
  const snippet = lines.join('\n');

  let guidance = 'Verify that this resource change is intentional before running apply.';
  if (item.is_stateful && item.action === 'replace') {
    guidance = 'Add prevent_destroy = true to protect stateful data, or replicate state before modifying immutable attributes.';
  } else if (item.action === 'replace') {
    guidance = 'Add create_before_destroy = true to provision the replacement instance prior to terminating the active workload.';
  } else if (item.action === 'delete') {
    guidance = 'Ensure all upstream and downstream dependencies have been successfully rerouted before approving deletion.';
  }

  return { snippet, guidance };
}
