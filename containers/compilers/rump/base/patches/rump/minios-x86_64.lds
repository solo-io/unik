OUTPUT_FORMAT("elf64-x86-64", "elf64-x86-64", "elf64-x86-64")
OUTPUT_ARCH(i386:x86-64)
ENTRY(_start)
SECTIONS
{
  . = 0x0;
  _text = .;			/* Text and read-only data */
  .text : {
	*(.text)
	*(.text.*)
	*(.gnu.warning)
	} = 0x9090

  _etext = .;			/* End of text section */

  .rodata : {
          *(.rodata)
	  *(.rodata.*)

          /* Constructors/destructors */
          . = ALIGN(64 / 8);
          __init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.init_array.*))
          *(SORT_BY_INIT_PRIORITY(.ctors*))
          *(.init_array)
          __init_array_end = . ;
  		
          __y1_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y1init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y1ctors*))
          *(.y1init_array)
          __y1_init_array_end = . ;
          
          __y2_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y2init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y2ctors*))
          *(.y2init_array)
          __y2_init_array_end = . ;
          
          __y3_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y3init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y3ctors*))
          *(.y3init_array)
          __y3_init_array_end = . ;
          
          __y4_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y4init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y4ctors*))
          *(.y4init_array)
          __y4_init_array_end = . ;
          
          __y5_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y5init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y5ctors*))
          *(.y5init_array)
          __y5_init_array_end = . ;
          
          __y6_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y6init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y6ctors*))
          *(.y6init_array)
          __y6_init_array_end = . ;
          
          __y7_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y7init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y7ctors*))
          *(.y7init_array)
          __y7_init_array_end = . ;
          
          __y8_init_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.y8init_array.*))
          *(SORT_BY_INIT_PRIORITY(.y8ctors*))
          *(.y8init_array)
          __y8_init_array_end = . ;
          
          __fini_array_start = . ;
          *(SORT_BY_INIT_PRIORITY(.fini_array.*))
          *(SORT_BY_INIT_PRIORITY(.dtors*))
          *(.fini_array)
          __fini_array_end = . ;
	  }

  . = ALIGN(4096);
  _erodata = .;

  .data : {			/* Data */
	*(.data)
	}

  _edata = .;			/* End of data section */

  .tdata : {
	_tdata_start = . ;
	*(.tdata)
	_tdata_end = . ;
	}

  .tbss : {
	_tbss_start = . ;
	*(.tbss)
	_tbss_end = . ;
	}

  __bss_start = .;		/* BSS */
  .bss : {
	*(.bss)
	*(.bss.*)
	}
  _end = . ;

  /* Sections to be discarded */
  /DISCARD/ : {
	*(.data.exit)
	*(.exitcall.exit)
	}

  /* Stabs debugging sections.  */
  .stab 0 : { *(.stab) }
  .stabstr 0 : { *(.stabstr) }
  .stab.excl 0 : { *(.stab.excl) }
  .stab.exclstr 0 : { *(.stab.exclstr) }
  .stab.index 0 : { *(.stab.index) }
  .stab.indexstr 0 : { *(.stab.indexstr) }
  .comment 0 : { *(.comment) }
}