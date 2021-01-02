package expense

// Reference: https://lists.gnucash.org/pipermail/gnucash-user/2014-December/057344.html
const Query = `
WITH RECURSIVE tree (
    guid,
    parent_guid,
    name,
    name_tree,
    name_tabs,
    account_type,
    depth
) AS (
    SELECT guid,
           parent_guid,
           name,
           '' || name AS name_tree,
           '' AS name_tabs,
           account_type,
           0 AS depth
    FROM accounts
    WHERE parent_guid IS NULL
    AND NAME <> 'Template Root'

    UNION ALL

    SELECT a.guid,
           a.parent_guid,
           a.name,
           tree.name_tree || ':' || a.name AS name_tree,
           substr('.................................',1,depth*2) || a.name AS name_tabs,
           a.account_type,
           depth + 1 AS depth
    FROM tree
    JOIN accounts a
    ON tree.guid = a.parent_guid
)

SELECT tr.account_type,
    tr.depth,
    tr.name AS depth_name,
    tr.name_tree AS full_name,
    (SELECT ROUND(SUM(CAST(s.value_num AS DOUBLE) / CAST (s.value_denom AS DOUBLE)), 2) AS value
     FROM tree tr1
     LEFT JOIN splits s
     ON tr1.guid = s.account_guid
     LEFT JOIN transactions tx
     ON s.tx_guid = tx.guid
     WHERE tr1.depth >= tr.depth
     AND tr1.account_type = tr.account_type
     AND tr1.name_tree LIKE tr.name_tree || '%'
     AND tx.post_date LIKE $1 || '%') AS value
FROM tree tr
WHERE tr.depth = $2
AND tr.account_type = 'EXPENSE'
AND value > 0
ORDER BY tr.name_tree, value DESC;
`

type Row struct {
	AccountType string  `csv:"account_type" db:"account_type"`
	Depth       int     `csv:"depth"        db:"depth"`
        DepthName   string  `csv:"depth_name"   db:"depth_name"`
	FullName    string  `csv:"full_name"    db:"full_name"`
	Value       float64 `csv:"value"        db:"value"`
}
