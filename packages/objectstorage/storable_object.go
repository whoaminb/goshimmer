package objectstorage

type StorableObject interface {
	GetId() []byte
	Update(object StorableObject)
	Marshal() ([]byte, error)
	Unmarshal(key []byte, serializedObject []byte) (StorableObject, error)
}
