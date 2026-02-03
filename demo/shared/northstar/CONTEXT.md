# Northstar Group Demo Context

This document provides background context for the Northstar Group demo scenarios. Use this narrative when presenting demos to maintain consistency and realism.

## Company Overview

**Northstar Group** is a fictional enterprise conglomerate with three main divisions:

| Division | Focus | Key Systems |
|----------|-------|-------------|
| **Northstar Logistics** | Global supply chain and warehousing | SAP S/4HANA, warehouse management, fleet tracking |
| **Northstar Digital** | E-commerce and digital platforms | Cloud-native apps, payment processing, customer portals |
| **Northstar Financial** | Payment processing and corporate finance | ERP, compliance systems, fraud detection |

The company has approximately 15,000 employees across North America and Europe, with headquarters in Chicago.

## Internal Tool: Northstar ToolsHub

**Northstar ToolsHub** is the company's internal platform feedback portal, built on Fider. It serves as the central place for employees across all divisions to:

- Submit feature requests for internal tools
- Vote on and discuss proposed improvements
- Track the status of requested features
- Provide feedback on deployed changes

The Platform Engineering team reviews submissions weekly during their Thursday stand-up and updates statuses accordingly.

## Personas

The demo data includes four realistic personas representing different roles and departments:

### Alex Rivera (Administrator)
- **Role**: Platform Engineering Lead
- **Email**: alex.rivera@northstar.io
- **Fider Role**: Administrator
- **Background**: 8 years at Northstar, leads the 12-person Platform Engineering team. Responsible for internal tooling strategy and prioritization. Former SRE who transitioned to platform engineering.
- **Demo Login**: `alex`

### Sarah Chen (Collaborator)
- **Role**: Product Manager, Digital Platforms
- **Email**: sarah.chen@northstar.io
- **Fider Role**: Collaborator
- **Background**: 4 years at Northstar Digital. Bridges the gap between business needs and technical implementation. Advocates for UX improvements and user-facing features.
- **Demo Login**: `sarah`

### Marcus Wright (Visitor)
- **Role**: Operations Analyst, Logistics
- **Email**: marcus.wright@northstar.io
- **Fider Role**: Visitor
- **Background**: 6 years in Northstar Logistics warehouse operations. Power user of internal tools, frequently encounters pain points in daily workflows. Champions requests that improve floor efficiency.
- **Demo Login**: `marcus`

### Jennifer Patel (Visitor)
- **Role**: Finance Systems Lead
- **Email**: jennifer.patel@northstar.io
- **Fider Role**: Visitor
- **Background**: 5 years in Northstar Financial. Focuses on compliance, audit trails, and integration with financial systems. Key stakeholder for SOX compliance features.
- **Demo Login**: `jennifer`

## Demo Login Mapping

When using the demo login feature (`/demo/login/<persona>`), use these slugs:

| Slug | User | Email |
|------|------|-------|
| `alex` | Alex Rivera | alex.rivera@northstar.io |
| `sarah` | Sarah Chen | sarah.chen@northstar.io |
| `marcus` | Marcus Wright | marcus.wright@northstar.io |
| `jennifer` | Jennifer Patel | jennifer.patel@northstar.io |

## Feature Request Themes

The seeded feature requests reflect realistic enterprise concerns:

### Security & Compliance
- Okta SSO session management
- Two-factor authentication (completed)
- Azure AD group sync
- Splunk audit log integration

### Integration
- SAP S/4HANA real-time sync
- PagerDuty alerting
- ServiceNow ticket creation
- SendGrid email (completed)

### User Experience
- Dark mode
- Global search (completed)
- Keyboard shortcuts (completed)
- Custom dashboard widgets

### Operations
- Mobile app stability
- API rate limiting visibility
- Bulk import capabilities

## Presentation Tips

1. **Start with Alex**: Log in as Alex to show the admin perspective and feature prioritization workflow.

2. **Switch to Marcus**: Demonstrate the end-user experience of submitting feedback and seeing responses.

3. **Highlight the workflow**: Show how a feature moves from Open → Planned → Started → Completed.

4. **Use realistic language**: Reference "the Thursday stand-up" or "InfoSec approval" to maintain the enterprise feel.

5. **Mention cross-division impact**: Features like SAP sync affect Logistics, while audit logging affects Finance.

## Technical Notes

- All users share the same tenant (`toolshub`)
- Session timeout and authentication are configured for demo mode
- Passwords are not used; demo login endpoint handles authentication
- The database can be reset to this seed state at any time
