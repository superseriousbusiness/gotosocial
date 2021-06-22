package pg

import (
	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) GetUserCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Account{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count where the domain field is null
		q = q.Where("? IS NULL", pg.Ident("domain"))
	} else {
		q = q.Where("domain = ?", domain)
	}

	// don't count the instance account or suspended users
	q = q.Where("username != ?", domain).Where("? IS NULL", pg.Ident("suspended_at"))

	return q.Count()
}

func (ps *postgresService) GetStatusCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Status{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count where local is true
		q = q.Where("local = ?", true)
	} else {
		// join on the domain of the account
		q = q.Join("JOIN accounts AS account ON account.id = status.account_id").
		Where("account.domain = ?", domain)
	}

	return q.Count()
}

func (ps *postgresService) GetDomainCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Instance{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count other instances it knows about
		// TODO: exclude domains that are blocked or silenced
		q = q.Where("domain != ?", domain)
	} else {
		// TODO: implement federated domain counting properly for remote domains
		return 0, nil
	}

	return q.Count()
}
