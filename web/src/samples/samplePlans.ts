import awsEcommerceData from './aws_ecommerce.json';
import awsFixedData from './aws_fixed.json';
import gcpMicroData from './gcp_micro.json';
import cleanInfraData from './clean_infra.json';
import type { AnalyzedPlan } from '../types';

export interface SamplePlanOption {
  id: string;
  title: string;
  subtitle: string;
  badge: string;
  badgeColor: string;
  plan: AnalyzedPlan;
}

export const SAMPLE_PLANS: SamplePlanOption[] = [
  {
    id: 'aws_ecommerce',
    title: 'AWS Production E-Commerce',
    subtitle: 'Stateful RDS replacement with downstream impact & open SSH ingress',
    badge: 'Critical Destruction',
    badgeColor: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
    plan: awsEcommerceData as unknown as AnalyzedPlan,
  },
  {
    id: 'aws_fixed',
    title: 'AWS E-Commerce (Remediated)',
    subtitle: 'Zero replacements, restricted SSH, and verified safe upgrade path',
    badge: 'Safe Remediated',
    badgeColor: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
    plan: awsFixedData as unknown as AnalyzedPlan,
  },
  {
    id: 'gcp_micro',
    title: 'GCP Microservices Platform',
    subtitle: 'Public firewall ingress 0.0.0.0/0 on RDP & publicly exposed Cloud SQL',
    badge: 'Security Violations',
    badgeColor: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
    plan: gcpMicroData as unknown as AnalyzedPlan,
  },
  {
    id: 'clean_infra',
    title: 'Clean Multi-Cloud Setup',
    subtitle: 'VPC networking with private subnets and restricted HTTPS rules',
    badge: 'Safe Plan',
    badgeColor: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
    plan: cleanInfraData as unknown as AnalyzedPlan,
  },
];
