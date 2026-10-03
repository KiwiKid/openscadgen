include <BOSL2/std.scad>;
/*
$fn = 100;

clipLength = 20;

clipHeight = 10;
clipWidth = 10;

lockingWedgeHeight = 6;

clipWallSize = 2;

module drawer_no_drop_locking_clip(){
difference(){
	cuboid([clipLength-0.01,10,clipHeight], anchor=BOT);
    
    up(clipWallSize)
    fwd(3)
    	
        cuboid([clipLength,2,clipHeight-clipWallSize], anchor=BOT);
    
    up(2)
    	cuboid([clipLength,5,2], anchor=BOT);
        
        
    up(clipWallSize*2.5)
    back(clipWallSize)
    	cuboid([clipLength,clipWidth-clipWallSize*2+0.01,clipHeight], anchor=BOT);
    }
    rotate([0,180,0])
    up(lockingWedgeHeight)
    wedge([clipLength,10, lockingWedgeHeight], anchor=TOP);
}

drawer_no_drop_locking_clip();

*/
include <BOSL2/std.scad>;

$fn = 100;

clipLength = 20;
lockingWedgeHeight = 14;
clipWallSize = 1.7;
clipWallBackSize = 3;

innerHoleHeight = 10;

tubeLength = 35;
tubeThickness = 5;

innerHoleWidth = 12;
innerHoleSize = [clipLength, innerHoleWidth, innerHoleHeight];

includeFrontSlot = true;

slotDepth = 22;
slotWidth = 3;
slotHeight = 20;
frontSlotSize = [slotWidth,slotDepth,slotHeight];
frontSlotDown = 3;
frontSlotMove = [0,0,frontSlotDown];

clipHeight = innerHoleHeight+clipWallSize*2+tubeLength;
clipWidth = innerHoleWidth+clipWallSize*2;


// Set your global chamfer size here
chamfSize = 3.5;
wedge_edges = [1, // Top face edges (including the top apex ridge), // Vertical/sloped side edges
    [0, 0, 0, 0]  // Bottom face edges (kept perfectly flat to mount flush)
];

debugMode = "sliced2";
debugSlice = [3,300,300];
debugSliceMove = [0,0,3];
wedgeOverhang = 12.1;

cutoutDown = 40;
angleCutoutDown = 0;
angleCutoutHeight = 3;
cutoutMove = [0,30,cutoutDown];


wedgeCutoutRotate = 65;

spinRadius = 2;
spineDown = 14;
spineMove = [0,tubeThickness/2-spinRadius*0.9,spineDown];

module drawer_no_drop_locking_clip(){
    intersection(){
    union(){
    
    difference(){
    
        // Passing chamfer directly to cuboid handles all edges/corners automatically
        cuboid([clipLength-0.01, clipWidth, clipHeight], chamfer=chamfSize, anchor=BOT,  edges="X", except_edges=BOT+FWD);

   
        
        // passage tube
        up(clipWallSize)
      //  fwd(clipWallSize)
      fwd(innerHoleWidth/2-tubeThickness/2)
      up(innerHoleHeight)
            cuboid([clipLength, tubeThickness, clipHeight-clipWallSize], anchor=BOT);
        
        // inner holding hole
        up(2)
            cuboid(innerHoleSize, anchor=BOT, chamfer=chamfSize/2, edges="X", except_edges=[TOP]);
            
            
            
        /// BOT WEDGE CUTOUT
      up(clipWallBackSize*2+innerHoleHeight)
        back(clipWallSize*1.5+clipWallSize+tubeThickness/1.6)
            cuboid([clipLength, clipWidth-clipWallSize*2+0.01, clipHeight], anchor=BOT);
        /*      move(cutoutMove)
            */
            
            
            
          // hexagon(r=20, anchor=CENTER);
          
          
            //     if(includeFrontSlot){
       
    //   up(frontSlotDown)
      // fwd(slotDepth/2)
     /*  move(frontSlotMove)
       left_half()
       linear_extrude(r=10)
      pentagon(r=20, anchor=CENTER, align_tip = FRONT);*/
    //       cuboid(frontSlotSize);
        //  }
    }
    
    // Fixed the empty rotate() from the original snippet which can cause warnings
rotate([0, 180, 0])
back(2)
  //  up(lockingWedgeHeight)
    fwd(clipWidth/2+wedgeOverhang)
    
    // A prismoid with a zero-width top face acts as a wedge, supporting full edge masking
   // difference(){
    
        /// TOP WEDGE
   /* prismoid(
        size1=[clipLength, 39], 
        size2=[clipLength, 25], 
        h=lockingWedgeHeight, 
        shift=[0, 0],
//        rounding=0.1,          // Change to chamfer=0.8 if you prefer flat chamfers
    //    edges=wedge_edges,     // Applies our edge override matrix
        anchor=TOP,
    );*/
    
    cuboid([clipLength,40,20], anchor=BOT);
    
    
        /// TOP WEDGE CUTOUT
      /*      xrot(90)
            back(8)
            down(cutoutDown)
            linear_extrude(h=100)
            pentagon(r = 25, align_tip = FRONT);*/
     //       }
    }
    
    union(){
    // Just the front of the wedge
    down(angleCutoutDown)
    xrot(-wedgeCutoutRotate)
        cuboid([clipLength-0.01,angleCutoutHeight, clipHeight*10], anchor=BACK);
        
        
        //    down(8)
   // xrot(40)
       cuboid([clipLength-0.01, 15, clipHeight], anchor=BOT);
       

        
        }
        
        //#cuboid([clipLength-0.01, 10, clipHeight], anchor=TOP+FWD);
    }
    
    
    //xrot(90)
            move(spineMove)
            linear_extrude(h=30)
            pentagon(r = spinRadius, align_tip = BACK);
}


if(debugMode == "sliced"){
intersection(){
drawer_no_drop_locking_clip();
move(debugSliceMove)
    cuboid(debugSlice);
}
} else {

drawer_no_drop_locking_clip();
}

