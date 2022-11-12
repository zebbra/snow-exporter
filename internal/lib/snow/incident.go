package snow

import (
	"context"
)

func (c *Client) Incident(ctx context.Context) ([]Incident, error) {
	resp, err := c.Fetch(
		ctx,
		"/api/now/table/incident?sysparm_limit=500&sysparm_query=active=true^ORDERBYDESCsys_created_on&sysparm_display_value=true&sysparm_exclude_reference_link=true&sysparm_fields=sys_id,number,state,priority,short_description,description,caller_id,opened_at,closed_at,resolved_by,location,service_offeringclose_notes,child_incidents,assigned_to,sys_updated_on",
		&IncidentList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*IncidentList)
	return list.Result, nil
}

type Incident struct {
	ShortDescription string `json:"short_description"`
	ClosedAt         string `json:"closed_at"`
	Description      string `json:"description"`
	Priority         string `json:"priority"`
	ChildIncidents   string `json:"child_incidents"`
	SysId            string `json:"sys_id"`
	Number           string `json:"number"`
	OpenedAt         string `json:"opened_at"`
	ResolvedBy       string `json:"resolved_by"`
	CallerId         string `json:"caller_id"`
	Location         string `json:"location"`
	State            string `json:"state"`
	AssignedTo       string `json:"assigned_to"`
	SysUpdatedOn     string `json:"sys_updated_on"`
}

func (i Incident) ModifiedAt() string {
	if i.SysUpdatedOn == "" {
		return i.OpenedAt
	}

	return i.SysUpdatedOn
}

type IncidentList struct {
	Result []Incident `json:"result"`
}
