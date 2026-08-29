---
title: Relevance preferences
description: How Jobmatcha filters and ranks the roles it finds.
---

Jobmatcha uses deterministic rules. Preferences control whether a role reaches the feed and how prominently it appears there.

## Filters

- **Location keywords** are a hard filter when you configure them: a role must match at least one location keyword in its title, department, or location.
- **Excluded keywords** remove roles when they appear in the title or department.
- **Work types** remove roles that explicitly advertise an unselected type. A role with no detected work-type signal remains visible.
- **Freshness** limits the feed to postings inside the selected age window, with an extra one-day tolerance. Roles without a posting date are excluded while this filter is enabled. Choose **Any date** to disable that filter.

## Ranking

Include-keyword matches increase a role's score. A title match carries more weight than a match in the department, location, or description. An explicitly detected selected work type adds a small bonus.

The displayed match percentage also applies a recency factor, so otherwise similar newer roles appear higher. Open a role to inspect its match analysis, including the matched fields, work-type bonus, and recency adjustment.

![Role details showing match analysis](/images/role-detail.png)

> [!NOTE]
> A role can be visible with a low or zero percentage when it passes your filters but does not contain an include-keyword match. Use include keywords to make the shortlist more selective.
