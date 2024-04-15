package foundation

type BitFlags struct {
	flags         uint32
	changeHandler func(flag uint32, currentValue bool)
}

func NewBitFlags() *BitFlags {
	return &BitFlags{}
}
func NewBitFlagsFromValue(value uint32) BitFlags {
	return BitFlags{flags: value}
}
func (f *BitFlags) Set(flag uint32) {
	f.flags |= flag
	f.onStateChanged(flag, true)
}

func (f *BitFlags) Unset(flag uint32) {
	f.flags &^= flag
	f.onStateChanged(flag, false)
}
func (f *BitFlags) IsSet(flag uint32) bool {
	return f.flags&flag != 0
}

func (f *BitFlags) Underlying() uint32 {
	return f.flags
}

func (f *BitFlags) SetOnChangeHandler(change func(flag uint32, currentValue bool)) {
	f.changeHandler = change
}

func (f *BitFlags) onStateChanged(flag uint32, currentValue bool) {
	if f.changeHandler != nil {
		f.changeHandler(flag, currentValue)
	}
}

func (f *BitFlags) Init(flags uint32) {
	f.flags = flags
}
