import React from 'react';
import { ShieldCheck, AlertTriangle, XCircle, Clock, ShieldAlert } from 'lucide-react';

export default function Badge({ status, label }) {
  const normalized = (status || '').toLowerCase();

  let icon = null;
  switch (normalized) {
    case 'active':
    case 'verified':
      icon = <ShieldCheck size={13} />;
      break;
    case 'expired':
      icon = <Clock size={13} />;
      break;
    case 'recalled':
      icon = <ShieldAlert size={13} />;
      break;
    case 'not_found':
    case 'invalid':
      icon = <XCircle size={13} />;
      break;
    case 'suspended':
      icon = <AlertTriangle size={13} />;
      break;
    default:
      icon = null;
  }

  const text = label || normalized.replace('_', ' ');

  return (
    <span className={`badge badge-${normalized}`}>
      {icon}
      {text}
    </span>
  );
}
