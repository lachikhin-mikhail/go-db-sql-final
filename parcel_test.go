package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

type TestSuite struct {
	suite.Suite
	db    *sql.DB
	store ParcelStore
}

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (suite *TestSuite) SetupSuite() {
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		return
	}
	suite.db = db
	suite.store = NewParcelStore(db)

}
func (suite *TestSuite) TearDownSuite() {
	err := suite.db.Close()
	require.NoError(suite.T(), err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func (suite *TestSuite) TestAddGetDelete() {
	// prepare
	parcel := getTestParcel()

	// add
	num, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), num)

	// get
	p, err := suite.store.Get(num)
	require.NoError(suite.T(), err)
	// Adjust for parcel default number being 0
	parcel.Number = num
	require.Equal(suite.T(), parcel, p)

	// delete
	err = suite.store.Delete(num)
	require.NoError(suite.T(), err)

	// get check
	_, err = suite.store.Get(num)
	require.Error(suite.T(), err)

}

// TestSetAddress проверяет обновление адреса
func (suite *TestSuite) TestSetAddress() {
	// prepare
	parcel := getTestParcel()

	// add
	num, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err)
	require.True(suite.T(), (num > 0))

	// set address
	newAddress := "new test address"
	err = suite.store.SetAddress(num, newAddress)
	require.NoError(suite.T(), err)

	// check
	p, err := suite.store.Get(num)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), p.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func (suite *TestSuite) TestSetStatus() {
	// prepare
	parcel := getTestParcel()

	// add
	num, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err)
	require.True(suite.T(), (num > 0))

	// set status
	err = suite.store.SetStatus(num, ParcelStatusSent)
	require.NoError(suite.T(), err)

	// check
	p, err := suite.store.Get(num)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ParcelStatusSent, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func (suite *TestSuite) TestGetByClient() {
	// prepare

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := suite.store.Add(parcels[i])
		require.NoError(suite.T(), err)
		require.True(suite.T(), (id > 0))
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
	}
	// Не совсем понял комментарий, по поводу наличия в testify функции для работы с массивами которой можно заменить данный цикл :')
	// Но так как это был precode, я полагаю что это был совет, а не требование, поэтому чтобы успеть закрыть проект до дедлайна отправляю так.

	// get by client
	storedParcels, err := suite.store.GetByClient(client)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(parcels), len(storedParcels))

	// check
	assert.ElementsMatch(suite.T(), parcels, storedParcels)
}
