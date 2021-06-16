package pg

import (
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"sync"
)

func (ps *postgresService) StatusParents(status *gtsmodel.Status) ([]*gtsmodel.Status, error) {
	parents := []*gtsmodel.Status{}
	ps.statusParent(status, &parents)

	return parents, nil
}

func (ps *postgresService) statusParent(status *gtsmodel.Status, foundStatuses *[]*gtsmodel.Status) {
	if status.InReplyToID == "" {
		return
	}

	parentStatus := &gtsmodel.Status{}
	if err := ps.conn.Model(parentStatus).Where("id = ?", status.InReplyToID).Select(); err == nil {
		*foundStatuses = append(*foundStatuses, parentStatus)
	}

	ps.statusParent(parentStatus, foundStatuses)
}

func (ps *postgresService) StatusChildren(status *gtsmodel.Status) ([]*gtsmodel.Status, error) {
	children := []*gtsmodel.Status{}
	// ps.statusChildren(status, &children)
	
	return children, nil
}

func (ps *postgresService) statusChildren(status *gtsmodel.Status, foundStatuses *sync.Map) {
	// immediateChildren := []*gtsmodel.Status{}
	
	// foundStatuses.Store()

	// err := ps.conn.Model(&immediateChildren).Where("in_reply_to_id = ?", status.ID).Select()
	// if err != nil {
	// 	return
	// }

	// for _, child := range immediateChildren {
	// 	f[""][0] = child
	// }

	return
}
