package main

import (
	"gorm.io/gorm"
	"time"
)

type DaoDevice interface {
	LoadOnlineDevices() (map[string]*Device, error)

	LoadDevices() (map[string]*Device, error)

	SaveDevice(device *Device) error

	RefreshHeartbeat(deviceId string, now time.Time, addr string) error

	QueryDevice(id string) (*Device, error)

	QueryDevices(page int, size int) ([]*Device, int, error)

	UpdateDeviceStatus(deviceId string, status OnlineStatus) error

	UpdateDeviceInfo(deviceId string, device *Device) error

	UpdateOfflineDevices(deviceIds []string) error

	ExistDevice(deviceId string) bool
}

type daoDevice struct {
}

func (d *daoDevice) LoadOnlineDevices() (map[string]*Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d *daoDevice) LoadDevices() (map[string]*Device, error) {
	var devices []*Device
	tx := db.Find(&devices)
	if tx.Error != nil {
		return nil, tx.Error
	}

	deviceMap := make(map[string]*Device)
	for _, device := range devices {
		deviceMap[device.DeviceID] = device
	}

	return deviceMap, nil
}

func (d *daoDevice) SaveDevice(device *Device) error {
	return DBTransaction(func(tx *gorm.DB) error {
		old := Device{}
		if db.Select("id").Where("device_id =?", device.DeviceID).Take(&old).Error == nil {
			device.ID = old.ID
		}

		if device.ID == 0 {
			//return tx.Create(&old).Error
			return tx.Save(device).Error
		} else {
			return tx.Model(device).Select("Transport", "RemoteAddr", "Status", "RegisterTime", "LastHeartbeat").Updates(*device).Error
		}
	})
}

func (d *daoDevice) UpdateDeviceInfo(deviceId string, device *Device) error {
	return DBTransaction(func(tx *gorm.DB) error {
		var condition = make(map[string]interface{})
		if device.Manufacturer != "" {
			condition["manufacturer"] = device.Manufacturer
		}
		if device.Model != "" {
			condition["model"] = device.Model
		}
		if device.Firmware != "" {
			condition["firmware"] = device.Firmware
		}
		if device.Name != "" {
			condition["name"] = device.Name
		}
		return tx.Model(&Device{}).Where("device_id =?", deviceId).Updates(condition).Error
	})
}

func (d *daoDevice) UpdateDeviceStatus(deviceId string, status OnlineStatus) error {
	return DBTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Device{}).Where("device_id =?", deviceId).Update("status", status).Error
	})
}

func (d *daoDevice) RefreshHeartbeat(deviceId string, now time.Time, addr string) error {
	if tx := db.Select("id").Take(&Device{}, "device_id =?", deviceId); tx.Error != nil {
		return tx.Error
	}
	return DBTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Device{}).Select("LastHeartbeat", "Status", "RemoteAddr").Where("device_id =?", deviceId).Updates(&Device{
			LastHeartbeat: now,
			Status:        ON,
			RemoteAddr:    addr,
		}).Error
	})
}

func (d *daoDevice) QueryDevice(id string) (*Device, error) {
	var device Device
	tx := db.Where("device_id =?", id).Take(&device)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &device, nil
}

func (d *daoDevice) QueryDevices(page int, size int) ([]*Device, int, error) {
	var devices []*Device
	tx := db.Limit(size).Offset((page - 1) * size).Find(&devices)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	var total int64
	tx = db.Model(&Device{}).Count(&total)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	for _, device := range devices {
		count, _ := ChannelDao.QueryChanelCount(device.DeviceID)
		online, _ := ChannelDao.QueryOnlineChanelCount(device.DeviceID)
		device.ChannelsOnline = online
		device.ChannelsTotal = count
	}

	return devices, int(total), nil
}

func (d *daoDevice) UpdateOfflineDevices(deviceIds []string) error {
	return DBTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Device{}).Where("device_id in ?", deviceIds).Update("status", OFF).Error
	})
}

func (d *daoDevice) ExistDevice(deviceId string) bool {
	var device Device
	tx := db.Select("id").Where("device_id =?", deviceId).Take(&device)
	if tx.Error != nil {
		return false
	}

	return true
}
