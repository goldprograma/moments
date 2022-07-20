package internal

import (
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// MediaDelete 删除关系
func MediaDelete(media *moment.Media, db *sqlx.DB) (err error) {
	sql := "delete from media  where main_id = ?"
	_, err = db.Exec(sql, media.MainID)
	return
}

// MediaGet 查询所有数据
func MediaGet(media *moment.Media, db *sqlx.DB) ([]*moment.Media, error) {
	sql := "select * from media where main_id = ?"
	medias := make([]*moment.Media, 0)
	err := db.Select(&medias, sql, media.MainID)
	return medias, err
}

// MediasInsert 批量插入媒体插入关系
func MediasInsert(Medias []*moment.Media, db *sqlx.DB) error {
	// if tx != nil {
	// 	_, err := tx.NamedExec("insert into media (main_id,seq,name,ext,thum,region,size,thum_size,hash,duration,height,width,create_at) "+
	// 		" values (:main_id,:seq,:name,:ext,:thum,:region,:size,:thum_size,:hash,:duration,:height,:width,:create_at)",
	// 		Medias)
	// 	return err
	// }
	_, err := db.NamedExec("insert into media (main_id,seq,name,ext,thum,region,size,thum_size,hash,duration,height,width,create_at) "+
		" values (:main_id,:seq,:name,:ext,:thum,:region,:size,:thum_size,:hash,:duration,:height,:width,:create_at)",
		Medias)
	return err
}
