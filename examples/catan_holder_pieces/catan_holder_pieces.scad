

	include <BOSL2/std.scad>;

include <BOSL2/joiners.scad>;
	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
    // 	renderType = "hex";
    
	renderType = "port-offset";    
    
    
    holderWidth= 6;
    
    floorHeight= 1.5;
    edgeHolderHeight =4;
    catanPieceSize = 91.5;
    
    innerHoleSize = 80;

    
        slotZRotate = -30;

    dovetailType = "male";
    dovetailWidth = 13.8;
    dovetailHeight = 5.2;
    
    holderHeight = 3;


portHolderWidth = 30;
    portWidth = 24.5;
    
    holderDepth = 15;
    holderOffset = 0;
    
    dovetailAngle = 34;
    
    maleOffset = 3.5;//4.1;
    
    femaleOffset = 3.5;//2.2;//2.9;
    
designFileName = "catan_holder_pieces";
dovetailType = "female";
name = "port-offset-female-left";
renderType = "port-offset";
slotMoveX = -8;
slotMoveY = 8;
slotZRotate = -30;
version = "v0.4";




    
    module main_frame(){
    tube(or=(catanPieceSize+holderWidth)/2, ir=innerHoleSize/2+5, h=3, $fn=6, rounding_fn=64, ichamfer2=3, teardrop=true, anchor=CENTER+DOWN);
        
    }


	module holder_side(dovetailType="male"){
        difference(){
        
      //  cuboid([holderDepth,portHolderWidth,holderHeight], anchor=LEFT);
        
        if(dovetailType == "female"){
            rotate([90,0,90])
           dovetail(dovetailType, slide=100, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle);
        
        }
        
        //move([holderDepth-holderOffset,0,0])
       // cuboid([holderDepth,portWidth,10]);
        }
         if(dovetailType == "male"){
         rotate([90,0,90])
            dovetail(dovetailType, slide=floorHeight, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle);
            }
	}
    
roadLength = 25.4;
        roadSize = [5,roadLength,3.6];
        roadUp= 3.25;
        roadOffset =4;
	module catan_holder_pieces(){
		difference(){
        
        union(){
        
        linear_extrude(height=2, scale=0)
        hexagon(d=catanPieceSize, anchor=CENTER);
        
//tube(or=(catanPieceSize+holderWidth)/2, ir=catanPieceSize/2, h=2.5, $fn=6, rounding_fn=64, teardrop=true, anchor=CENTER+DOWN);
        
        main_frame();
        
        linear_extrude(h = edgeHolderHeight)
		hexagon(d=catanPieceSize+holderWidth, anchor=CENTER);
        }
        
        up(floorHeight)
         //cylindrical_extrude(ir=1, or=2)

        linear_extrude(h = edgeHolderHeight+10)
        hexagon(d=catanPieceSize,  anchor=CENTER);
        
        
        down(1)
       linear_extrude(h = 20)
		hexagon(d=innerHoleSize,  anchor=CENTER);
        
        
        for(i = [0:6]){
        if(i % 2 != 0){
            rotate([0,0,i*60+30])
            move([catanPieceSize/2-femaleOffset,0,0])
        holder_side(dovetailType="female");
        
        }
        
        up(roadUp)
            rotate([0,0,i*60+30])
            move([catanPieceSize/2-roadOffset,0,0])
        cuboid(roadSize);
        }
        }
        
        
        for(i = [0:6]){
        if(i % 2 == 0){
        
            rotate([0,0,i*60+30])
            move([catanPieceSize/2-maleOffset,0,floorHeight/2])
            holder_side(dovetailType="male");
        
        
        }
        }
        
	}
    
    portHolderHeight = 3;
    portOuterWidth = 27;
    portOuterHeight  = 26.8;
    portFrameWidth = portOuterWidth-3;
    portFrameHeight = portOuterHeight-6;
    portSlotWidth = portOuterWidth - 1;
portSlotHeight = 1.6;
portSlotHolderWidth = 12;

slotOffset=2;
    
    module catan_holder_port(dovetailType="male"){
    


        difference(){
        
               cuboid([portOuterWidth, portOuterHeight, portHolderHeight], anchor=LEFT+BOTTOM);
                
               
            union(){
              right(slotOffset+5)
                up(1)
                right(1)
                cuboid([portFrameWidth, portFrameHeight, portHolderHeight+0.01], anchor=LEFT+BOTTOM);
                

                right(slotOffset+5)
                up(0.4)
                cuboid([portSlotWidth, portSlotWidth, portSlotHeight], anchor=LEFT+BOTTOM);
                }
                
                
          if(dovetailType == "female"){
            rotate([90,0,-90])
         //   move([0,0,femalePortOffset])
            dovetail(dovetailType, slide=100, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle, anchor=BOTTOM+FWD);
        
        }
        
        
            up(roadUp)
        rotate([0,0,0])
        //move([catanPieceSize/2-roadOffset,0,0])
        cuboid(roadSize);
        
        }
        
        //move([holderDepth-holderOffset,0,0])
       // cuboid([holderDepth,portWidth,10]);
        
         if(dovetailType == "male"){
         rotate([90,0,-90])
         
          //  move([0,0,malePortOffset])
            dovetail(dovetailType, slide=floorHeight, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle, anchor=BOTTOM+FWD);
            }
        }
        
        
      
       module catan_holder_port_offset(dovetailType="male"){


        slotRotate = [0,0,slotZRotate];
        slotMoveX = 8;
        slotMoveY = 8;
        slotMove = [slotMoveX,slotMoveY,0];


        difference(){
               
               union(){
               cuboid([portOuterWidth-2, portOuterHeight, portHolderHeight], anchor=LEFT+BOTTOM);
               
                move(slotMove)
            rotate(slotRotate)
               cuboid([portOuterWidth, portOuterHeight, portHolderHeight], rounding=10, edges = "Z", except=[TOP,RIGHT,BOTTOM], anchor=LEFT+BOTTOM);
               }
               
                
        move(slotMove)
        rotate(slotRotate)
            union(){
               
               
              right(slotOffset+5)
                up(1)
                right(1)
                
                cuboid([portFrameWidth, portFrameHeight, portHolderHeight+0.01], anchor=LEFT+BOTTOM);
                

                right(slotOffset+6)
                up(0.4)
                cuboid([portSlotWidth, portSlotWidth, portSlotHeight], anchor=LEFT+BOTTOM);
                }
                
                
          if(dovetailType == "female"){
            rotate([90,0,-90])
         //   move([0,0,femalePortOffset])
            dovetail(dovetailType, slide=100, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle, anchor=BOTTOM+FWD);
        
        }
        
            up(roadUp)
        rotate([0,0,0])
        //move([catanPieceSize/2-roadOffset,0,0])
        cuboid(roadSize);
        
        }
        
        //move([holderDepth-holderOffset,0,0])
       // cuboid([holderDepth,portWidth,10]);
        
         if(dovetailType == "male"){
         rotate([90,0,-90])
         
          //  move([0,0,malePortOffset])
            dovetail(dovetailType, slide=floorHeight, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle, anchor=BOTTOM+FWD);
            }
        }

        
        
        
        

    sliced(renderType=renderType) {
        if(renderType == "hex"){
            catan_holder_pieces();
        } 
        
        if(renderType == "port"){
         //   if(dovetailType == "female"){
            catan_holder_port(dovetailType=dovetailType);
       /*     }
            if(dovetailType == "male"){
                catan_holder_port(dovetailType="male");
            }*/

               }
               
               
        if(renderType == "port-offset"){
            catan_holder_port_offset(dovetailType=dovetailType);
        }
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

