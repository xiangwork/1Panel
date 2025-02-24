package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	pgclient "github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"
)

func (u *BackupService) PostgresqlBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}

	timeNow := time.Now().Format("20060102150405")
	targetDir := path.Join(localDir, fmt.Sprintf("database/%s/%s/%s", req.Type, req.Name, req.DetailName))
	fileName := fmt.Sprintf("%s_%s.sql.gz", req.DetailName, timeNow)

	if err := handlePostgresqlBackup(req.Name, req.DetailName, targetDir, fileName); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:       req.Type,
		Name:       req.Name,
		DetailName: req.DetailName,
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    targetDir,
		FileName:   fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}
	return nil
}
func (u *BackupService) PostgresqlRecover(req dto.CommonRecover) error {
	if err := handlePostgresqlRecover(req, false); err != nil {
		return err
	}
	return nil
}

func (u *BackupService) PostgresqlRecoverByUpload(req dto.CommonRecover) error {
	file := req.File
	fileName := path.Base(req.File)
	if strings.HasSuffix(fileName, ".tar.gz") {
		fileNameItem := time.Now().Format("20060102150405")
		dstDir := fmt.Sprintf("%s/%s", path.Dir(req.File), fileNameItem)
		if _, err := os.Stat(dstDir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", dstDir, err)
			}
		}
		if err := handleUnTar(req.File, dstDir); err != nil {
			_ = os.RemoveAll(dstDir)
			return err
		}
		global.LOG.Infof("decompress file %s successful, now start to check test.sql is exist", req.File)
		hasTestSql := false
		_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "test.sql" {
				hasTestSql = true
				file = path
				fileName = "test.sql"
			}
			return nil
		})
		if !hasTestSql {
			_ = os.RemoveAll(dstDir)
			return fmt.Errorf("no such file named test.sql in %s", fileName)
		}
		defer func() {
			_ = os.RemoveAll(dstDir)
		}()
	}

	req.File = path.Dir(file) + "/" + fileName
	if err := handlePostgresqlRecover(req, false); err != nil {
		return err
	}
	global.LOG.Info("recover from uploads successful!")
	return nil
}
func handlePostgresqlBackup(database, dbName, targetDir, fileName string) error {
	cli, err := LoadPostgresqlClientByFrom(database)
	if err != nil {
		return err
	}
	defer cli.Close()

	backupInfo := pgclient.BackupInfo{
		Name:      dbName,
		TargetDir: targetDir,
		FileName:  fileName,

		Timeout: 300,
	}
	if err := cli.Backup(backupInfo); err != nil {
		return err
	}
	return nil
}

func handlePostgresqlRecover(req dto.CommonRecover, isRollback bool) error {
	isOk := false
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return buserr.WithName("ErrFileNotFound", req.File)
	}
	dbInfo, err := postgresqlRepo.Get(commonRepo.WithByName(req.DetailName), postgresqlRepo.WithByPostgresqlName(req.Name))
	if err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Name)
	if err != nil {
		return err
	}
	defer cli.Close()

	if !isRollback {
		rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("database/%s/%s_%s.sql.gz", req.Type, req.DetailName, time.Now().Format("20060102150405")))
		if err := cli.Backup(client.BackupInfo{
			Name:      req.DetailName,
			TargetDir: path.Dir(rollbackFile),
			FileName:  path.Base(rollbackFile),

			Timeout: 300,
		}); err != nil {
			return fmt.Errorf("backup postgresql db %s for rollback before recover failed, err: %v", req.DetailName, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := cli.Recover(client.RecoverInfo{
					Name:       req.DetailName,
					SourceFile: rollbackFile,

					Timeout: 300,
				}); err != nil {
					global.LOG.Errorf("rollback postgresql db %s from %s failed, err: %v", req.DetailName, rollbackFile, err)
				}
				global.LOG.Infof("rollback postgresql db %s from %s successful", req.DetailName, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}
	if err := cli.Recover(client.RecoverInfo{
		Name:       req.DetailName,
		SourceFile: req.File,
		Username:   dbInfo.Username,
		Timeout:    300,
	}); err != nil {
		return err
	}
	isOk = true
	return nil
}
