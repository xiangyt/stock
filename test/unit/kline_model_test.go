package unit

import (
	"reflect"
	"testing"

	"stock/internal/model"
)

func TestKLineModelStructure(t *testing.T) {
	t.Run("TestDailyDataStructure", func(t *testing.T) {
		var dailyData model.DailyData

		// 检查结构体字段
		v := reflect.ValueOf(dailyData)
		typ := reflect.TypeOf(dailyData)

		// 验证没有ID字段
		_, hasID := typ.FieldByName("ID")
		if hasID {
			t.Error("DailyData不应该有ID字段")
		}

		// 验证有TsCode字段
		tsCodeField, hasTsCode := typ.FieldByName("TsCode")
		if !hasTsCode {
			t.Error("DailyData应该有TsCode字段")
		} else {
			// 检查GORM标签
			gormTag := tsCodeField.Tag.Get("gorm")
			if gormTag == "" {
				t.Error("TsCode字段应该有gorm标签")
			}
			t.Logf("TsCode字段的gorm标签: %s", gormTag)
		}

		// 验证有TradeDate字段
		tradeDateField, hasTradeDate := typ.FieldByName("TradeDate")
		if !hasTradeDate {
			t.Error("DailyData应该有TradeDate字段")
		} else {
			// 检查GORM标签
			gormTag := tradeDateField.Tag.Get("gorm")
			if gormTag == "" {
				t.Error("TradeDate字段应该有gorm标签")
			}
			t.Logf("TradeDate字段的gorm标签: %s", gormTag)
		}

		// 验证字段数量
		numFields := v.NumField()
		t.Logf("DailyData结构体字段数量: %d", numFields)

		// 列出所有字段
		for i := 0; i < numFields; i++ {
			field := typ.Field(i)
			t.Logf("字段 %d: %s (类型: %s, 标签: %s)", i, field.Name, field.Type, field.Tag)
		}
	})

	t.Run("TestWeeklyDataStructure", func(t *testing.T) {
		var weeklyData model.WeeklyData
		typ := reflect.TypeOf(weeklyData)

		// 验证没有ID字段
		_, hasID := typ.FieldByName("ID")
		if hasID {
			t.Error("WeeklyData不应该有ID字段")
		}

		// 验证有TsCode和TradeDate字段
		_, hasTsCode := typ.FieldByName("TsCode")
		_, hasTradeDate := typ.FieldByName("TradeDate")

		if !hasTsCode {
			t.Error("WeeklyData应该有TsCode字段")
		}
		if !hasTradeDate {
			t.Error("WeeklyData应该有TradeDate字段")
		}

		t.Logf("WeeklyData结构验证通过")
	})

	t.Run("TestMonthlyDataStructure", func(t *testing.T) {
		var monthlyData model.MonthlyData
		typ := reflect.TypeOf(monthlyData)

		// 验证没有ID字段
		_, hasID := typ.FieldByName("ID")
		if hasID {
			t.Error("MonthlyData不应该有ID字段")
		}

		// 验证有TsCode和TradeDate字段
		_, hasTsCode := typ.FieldByName("TsCode")
		_, hasTradeDate := typ.FieldByName("TradeDate")

		if !hasTsCode {
			t.Error("MonthlyData应该有TsCode字段")
		}
		if !hasTradeDate {
			t.Error("MonthlyData应该有TradeDate字段")
		}

		t.Logf("MonthlyData结构验证通过")
	})

	t.Run("TestYearlyDataStructure", func(t *testing.T) {
		var yearlyData model.YearlyData
		typ := reflect.TypeOf(yearlyData)

		// 验证没有ID字段
		_, hasID := typ.FieldByName("ID")
		if hasID {
			t.Error("YearlyData不应该有ID字段")
		}

		// 验证有TsCode和TradeDate字段
		_, hasTsCode := typ.FieldByName("TsCode")
		_, hasTradeDate := typ.FieldByName("TradeDate")

		if !hasTsCode {
			t.Error("YearlyData应该有TsCode字段")
		}
		if !hasTradeDate {
			t.Error("YearlyData应该有TradeDate字段")
		}

		t.Logf("YearlyData结构验证通过")
	})
}

func TestKLineModelTags(t *testing.T) {
	t.Run("TestGormTags", func(t *testing.T) {
		// 测试DailyData的GORM标签
		typ := reflect.TypeOf(model.DailyData{})

		tsCodeField, _ := typ.FieldByName("TsCode")
		tradeDateField, _ := typ.FieldByName("TradeDate")

		tsCodeGormTag := tsCodeField.Tag.Get("gorm")
		tradeDateGormTag := tradeDateField.Tag.Get("gorm")

		t.Logf("TsCode GORM标签: %s", tsCodeGormTag)
		t.Logf("TradeDate GORM标签: %s", tradeDateGormTag)

		// 验证主键标签
		if tsCodeGormTag == "" {
			t.Error("TsCode字段应该有gorm标签")
		}
		if tradeDateGormTag == "" {
			t.Error("TradeDate字段应该有gorm标签")
		}
	})
}

func TestKLineModelInstantiation(t *testing.T) {
	t.Run("TestCreateInstances", func(t *testing.T) {
		// 测试创建实例
		dailyData := model.DailyData{
			TsCode:    "000001.SZ",
			TradeDate: 20250917,
			Open:      10.50,
			High:      11.20,
			Low:       10.30,
			Close:     11.00,
			Volume:    1000000,
			Amount:    10800000.00,
		}

		if dailyData.TsCode != "000001.SZ" {
			t.Error("TsCode字段设置失败")
		}
		if dailyData.TradeDate != 20250917 {
			t.Error("TradeDate字段设置失败")
		}

		t.Logf("DailyData实例创建成功: %+v", dailyData)

		// 测试其他K线数据类型
		weeklyData := model.WeeklyData{
			TsCode:    "000002.SZ",
			TradeDate: 20250917,
			Close:     15.50,
		}

		monthlyData := model.MonthlyData{
			TsCode:    "000003.SZ",
			TradeDate: 20250930,
			Close:     20.50,
		}

		yearlyData := model.YearlyData{
			TsCode:    "600000.SH",
			TradeDate: 20251231,
			Close:     30.50,
		}

		t.Logf("所有K线数据类型实例创建成功")
		t.Logf("WeeklyData: TsCode=%s, TradeDate=%d", weeklyData.TsCode, weeklyData.TradeDate)
		t.Logf("MonthlyData: TsCode=%s, TradeDate=%d", monthlyData.TsCode, monthlyData.TradeDate)
		t.Logf("YearlyData: TsCode=%s, TradeDate=%d", yearlyData.TsCode, yearlyData.TradeDate)
	})
}
