package objectstorage

type StorableObject interface {
	GetId() []byte
	Update(other StorableObject)
	Marshal() ([]byte, error)
	Unmarshal(key []byte, serializedObject []byte) (StorableObject, error)
}
